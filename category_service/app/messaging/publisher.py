import json
import logging
import threading
import time

import pika

from config import Config


class CategoryEventsPublisher:
    def __init__(self, config: Config, logger: logging.Logger) -> None:
        self.config = config
        self.logger = logger
        self._lock = threading.Lock()
        self._connection = None
        self._channel = None

    def publish_category_updated(
        self,
        *,
        category_uuid: str,
        workspace_id: str,
        actor_user_uuid: str | None,
        name: str | None,
    ) -> None:
        if not self.config.RABBITMQ_ENABLED or not name:
            return

        envelope = {
            "event_id": f"category-{category_uuid}-{int(time.time() * 1000)}",
            "event_type": "category.updated",
            "event_version": 1,
            "occurred_at": int(time.time()),
            "producer": "category_service",
            "payload": {
                "category_uuid": category_uuid,
                "workspace_id": workspace_id,
                "actor_user_uuid": actor_user_uuid or "",
                "name": name,
            },
        }
        self._publish("category.updated", envelope)

    def close(self) -> None:
        with self._lock:
            if self._channel and self._channel.is_open:
                try:
                    self._channel.close()
                except Exception:
                    self.logger.debug("failed to close category publisher channel", exc_info=True)
            if self._connection and self._connection.is_open:
                try:
                    self._connection.close()
                except Exception:
                    self.logger.debug("failed to close category publisher connection", exc_info=True)
            self._channel = None
            self._connection = None

    def _publish(self, routing_key: str, payload: dict) -> None:
        self._ensure_connection()
        body = json.dumps(payload).encode("utf-8")

        with self._lock:
            self._channel.basic_publish(
                exchange=self.config.RABBITMQ_EXCHANGE,
                routing_key=routing_key,
                body=body,
                properties=pika.BasicProperties(
                    content_type="application/json",
                    delivery_mode=2,
                    message_id=payload["event_id"],
                    type=payload["event_type"],
                    timestamp=payload["occurred_at"],
                ),
            )

    def _ensure_connection(self) -> None:
        with self._lock:
            if self._connection and self._connection.is_open and self._channel and self._channel.is_open:
                return

            if self._channel and self._channel.is_open:
                self._channel.close()
            if self._connection and self._connection.is_open:
                self._connection.close()

            self._connection = pika.BlockingConnection(pika.URLParameters(self.config.RABBITMQ_URL))
            self._channel = self._connection.channel()
            self._channel.exchange_declare(
                exchange=self.config.RABBITMQ_EXCHANGE,
                exchange_type="topic",
                durable=True,
            )
            self.logger.info("category events publisher connected")
