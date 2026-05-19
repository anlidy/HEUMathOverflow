from __future__ import annotations

import json
import logging
import threading
import time

import pika

from rag_service.chunking import ChunkBuilder
from rag_service.config import RabbitMQConfig
from rag_service.models import RagEvent
from rag_service.qdrant_store import QdrantIndex

LOGGER = logging.getLogger(__name__)


class RagEventConsumer:
    def __init__(self, config: RabbitMQConfig, store: QdrantIndex, chunk_builder: ChunkBuilder) -> None:
        self._config = config
        self._store = store
        self._chunk_builder = chunk_builder
        self._stop = threading.Event()
        self._threads: list[threading.Thread] = []

    def start(self) -> None:
        self._ensure_dead_letter_queue()
        for shard_id in range(self._config.rag_worker_count):
            thread = threading.Thread(target=self._consume_shard, args=(shard_id,), daemon=True)
            thread.start()
            self._threads.append(thread)

    def stop(self) -> None:
        self._stop.set()
        for thread in self._threads:
            thread.join(timeout=3)

    def _ensure_dead_letter_queue(self) -> None:
        connection = pika.BlockingConnection(pika.URLParameters(self._config.url))
        channel = connection.channel()
        channel.exchange_declare(
            exchange=self._config.dead_letter_exchange,
            exchange_type="direct",
            durable=True,
        )
        channel.queue_declare(queue=self._config.dead_letter_queue, durable=True)
        channel.queue_bind(
            queue=self._config.dead_letter_queue,
            exchange=self._config.dead_letter_exchange,
            routing_key=self._config.dead_letter_queue,
        )
        channel.close()
        connection.close()

    def _consume_shard(self, shard_id: int) -> None:
        queue_name = f"{self._config.queue_prefix}.shard-{shard_id}"
        binding_key = f"rag.data.*.{shard_id}"

        while not self._stop.is_set():
            try:
                connection = pika.BlockingConnection(pika.URLParameters(self._config.url))
                channel = connection.channel()
                channel.basic_qos(prefetch_count=1)
                channel.exchange_declare(
                    exchange=self._config.exchange,
                    exchange_type=self._config.exchange_type,
                    durable=True,
                )
                channel.queue_declare(
                    queue=queue_name,
                    durable=True,
                    arguments={
                        "x-dead-letter-exchange": self._config.dead_letter_exchange,
                        "x-dead-letter-routing-key": self._config.dead_letter_queue,
                    },
                )
                channel.queue_bind(
                    queue=queue_name,
                    exchange=self._config.exchange,
                    routing_key=binding_key,
                )
                LOGGER.info("rag consumer shard started: shard=%s queue=%s", shard_id, queue_name)

                def on_message(ch: pika.adapters.blocking_connection.BlockingChannel, method, properties, body: bytes) -> None:  # noqa: ANN001
                    self._handle_message(ch, method.delivery_tag, body)

                channel.basic_consume(queue=queue_name, on_message_callback=on_message, auto_ack=False)
                while not self._stop.is_set():
                    connection.process_data_events(time_limit=1)
                channel.close()
                connection.close()
            except Exception as exc:  # noqa: BLE001
                LOGGER.exception("rag consumer shard reconnect: shard=%s error=%s", shard_id, exc)
                time.sleep(1)

    def _handle_message(
        self,
        channel: pika.adapters.blocking_connection.BlockingChannel,
        delivery_tag: int,
        body: bytes,
    ) -> None:
        try:
            raw = json.loads(body)
            event = RagEvent.model_validate(raw)
            if event.type == "rag.data.deleted":
                self._store.delete_post(event.payload.post_id)
            else:
                records = self._chunk_builder.build(event.payload)
                self._store.replace_post(records)
            channel.basic_ack(delivery_tag=delivery_tag)
        except Exception as exc:  # noqa: BLE001
            LOGGER.exception("handle rag event failed: %s", exc)
            channel.basic_nack(delivery_tag=delivery_tag, requeue=False)
