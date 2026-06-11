import json
import logging
import threading
import time
from typing import Callable

import pika

from config import Config
from dao.model.dto import CreateGraphNoteDTO, UpdateGraphNoteDTO
from exceptions import AppException, NotFoundException
from service import CategoryService


CATEGORY_QUEUE_NAME = "category.domain-events"
ROUTING_KEYS = [
    "note.created",
    "note.updated",
    "note.deleted",
    "workspace.invite.accepted",
]


class DomainEventsConsumer:
    def __init__(
        self,
        config: Config,
        logger: logging.Logger,
        category_service_factory: Callable[[], CategoryService],
    ) -> None:
        self.config = config
        self.logger = logger
        self.category_service_factory = category_service_factory
        self._stop_event = threading.Event()
        self._thread: threading.Thread | None = None
        self._connection = None
        self._channel = None

    def start(self) -> None:
        if not self.config.RABBITMQ_ENABLED:
            self.logger.info("rabbitmq consumer disabled")
            return

        if self._thread and self._thread.is_alive():
            return

        self._thread = threading.Thread(target=self._run, name="category-domain-consumer", daemon=True)
        self._thread.start()
        self.logger.info("category domain events consumer thread started")

    def stop(self) -> None:
        self._stop_event.set()

        if self._channel and getattr(self._channel, "is_open", False):
            try:
                self._connection.add_callback_threadsafe(self._channel.stop_consuming)
            except Exception:
                self.logger.debug("failed to stop rabbitmq consumer gracefully", exc_info=True)

        if self._thread and self._thread.is_alive():
            self._thread.join(timeout=5)

        self._close()

    def _run(self) -> None:
        while not self._stop_event.is_set():
            try:
                self._consume()
            except Exception:
                self.logger.exception("category domain events consumer failed")
                self._close()
                if not self._stop_event.is_set():
                    time.sleep(5)

    def _consume(self) -> None:
        params = pika.URLParameters(self.config.RABBITMQ_URL)
        self._connection = pika.BlockingConnection(params)
        self._channel = self._connection.channel()
        self._channel.exchange_declare(
            exchange=self.config.RABBITMQ_EXCHANGE,
            exchange_type="topic",
            durable=True,
        )
        self._channel.queue_declare(queue=CATEGORY_QUEUE_NAME, durable=True)
        for routing_key in ROUTING_KEYS:
            self._channel.queue_bind(
                exchange=self.config.RABBITMQ_EXCHANGE,
                queue=CATEGORY_QUEUE_NAME,
                routing_key=routing_key,
            )
        self._channel.basic_qos(prefetch_count=10)

        self.logger.info("category domain events consumer connected")

        for method, _, body in self._channel.consume(
            CATEGORY_QUEUE_NAME,
            inactivity_timeout=1,
            auto_ack=False,
        ):
            if self._stop_event.is_set():
                break
            if method is None:
                continue

            delivery_tag = method.delivery_tag
            try:
                self._handle_message(body)
                self._channel.basic_ack(delivery_tag=delivery_tag)
            except NotFoundException:
                self.logger.warning("domain event skipped because graph target is absent", exc_info=True)
                self._channel.basic_ack(delivery_tag=delivery_tag)
            except AppException:
                self.logger.warning("domain event rejected by category service", exc_info=True)
                self._channel.basic_ack(delivery_tag=delivery_tag)
            except Exception:
                self.logger.exception("failed to process category domain event")
                self._channel.basic_ack(delivery_tag=delivery_tag)

    def _handle_message(self, body: bytes) -> None:
        payload = json.loads(body.decode("utf-8"))
        event_type = payload.get("event_type")
        event_payload = payload.get("payload") or {}

        if event_type == "note.created":
            self._handle_note_created(event_payload)
            return
        if event_type == "note.updated":
            self._handle_note_updated(event_payload)
            return
        if event_type == "note.deleted":
            self._handle_note_deleted(event_payload)
            return
        if event_type == "workspace.invite.accepted":
            self._handle_workspace_invite_accepted(event_payload)

    def _handle_note_created(self, event_payload: dict) -> None:
        dto = CreateGraphNoteDTO(
            uuid=event_payload["note_uuid"],
            workspace_id=event_payload["workspace_id"],
            workspace_name=event_payload.get("workspace_name"),
            workspace_type=event_payload.get("workspace_type"),
            author_user_uuid=event_payload.get("author_user_uuid") or event_payload.get("user_uuid"),
            category_uuid=event_payload["category_uuid"],
            header=event_payload["header"],
            created_date=event_payload["created_at"],
        )
        self.category_service_factory().create_note_node(note=dto)

    def _handle_note_updated(self, event_payload: dict) -> None:
        dto = UpdateGraphNoteDTO(
            workspace_id=event_payload["workspace_id"],
            category_uuid=event_payload.get("category_uuid"),
            header=event_payload.get("header"),
        )
        self.category_service_factory().update_note_node(note_uuid=event_payload["note_uuid"], note=dto)

    def _handle_note_deleted(self, event_payload: dict) -> None:
        self.category_service_factory().delete_note_node(
            note_uuid=event_payload["note_uuid"],
            workspace_id=event_payload["workspace_id"],
        )

    def _handle_workspace_invite_accepted(self, event_payload: dict) -> None:
        self.category_service_factory().ensure_workspace_member(
            workspace_id=event_payload["workspace_id"],
            user_uuid=event_payload["user_uuid"],
            workspace_name=event_payload.get("workspace_name"),
            workspace_type=event_payload.get("workspace_type"),
        )

    def _close(self) -> None:
        if self._channel and getattr(self._channel, "is_open", False):
            try:
                self._channel.close()
            except Exception:
                self.logger.debug("failed to close rabbitmq channel", exc_info=True)
        if self._connection and getattr(self._connection, "is_open", False):
            try:
                self._connection.close()
            except Exception:
                self.logger.debug("failed to close rabbitmq connection", exc_info=True)
        self._channel = None
        self._connection = None
