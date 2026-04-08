FROM python:3.13-slim-bookworm
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /app
RUN pip install --no-cache-dir "redis>=5.0" "websockets>=14.0"
COPY ws_server.py .
CMD ["python", "ws_server.py"]
