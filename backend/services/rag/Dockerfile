FROM python:3.12-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

COPY services/rag/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

COPY services/rag/main.py /app/main.py
COPY services/rag/config.yaml /app/config.yaml
COPY services/rag/README.md /app/README.md
COPY services/rag/rag_service /app/rag_service

EXPOSE 9091

CMD ["python", "main.py"]
