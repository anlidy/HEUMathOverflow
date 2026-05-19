FROM python:3.12-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_DEFAULT_TIMEOUT=1000 \
    PIP_PROGRESS_BAR=off \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple \
    PIP_EXTRA_INDEX_URL=https://download.pytorch.org/whl/cpu

COPY services/rag/model_server/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir --retries 20 torch==2.7.0+cpu && \
    pip install --no-cache-dir --retries 20 -r /app/requirements.txt

COPY services/rag/model_server/app.py /app/app.py

EXPOSE 8000

CMD ["python", "app.py"]
