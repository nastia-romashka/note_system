import os
from typing import Any

import yaml

from constants import CONFIG_FILE_PATH


class Config:
    __slots__ = [
        "_path",
        "DEBUG",
        "LOG_LEVEL",
        "LOGGER_NAME",
        "APP_HOST",
        "APP_PORT",
        "APP_TITLE",
        "API_PREFIX",
        "CORS_ALLOWED_ORIGINS",
        "NEO4J_SCHEME",
        "NEO4J_HOSTNAME",
        "NEO4J_PORT",
        "NEO4J_LOGIN",
        "NEO4J_PASSWORD",
        "NEO4J_DATABASE",
        "RABBITMQ_ENABLED",
        "RABBITMQ_URL",
        "RABBITMQ_EXCHANGE",
        "RABBITMQ_NOTE_UPDATED_QUEUE",
        "RABBITMQ_NOTE_UPDATED_ROUTING_KEY",
    ]

    def __init__(self, yaml_file: str) -> None:
        self._path = yaml_file or CONFIG_FILE_PATH
        self._set_defaults()
        self._read()
        self._read_env()

    def to_dict(self) -> dict[str, Any]:
        return {
            field.lower(): getattr(self, field)
            for field in self.__slots__
            if field != "_path"
        }

    @property
    def neo4j_uri(self) -> str:
        return f"{self.NEO4J_SCHEME}://{self.NEO4J_HOSTNAME}:{self.NEO4J_PORT}"

    def _set_defaults(self) -> None:
        self.DEBUG = True
        self.LOG_LEVEL = "DEBUG"
        self.LOGGER_NAME = "category_service"
        self.APP_HOST = "0.0.0.0"
        self.APP_PORT = 8081
        self.APP_TITLE = "Category Service"
        self.API_PREFIX = "/api"
        self.CORS_ALLOWED_ORIGINS = ["http://localhost:3000", "http://localhost:5173"]
        self.NEO4J_SCHEME = "bolt"
        self.NEO4J_HOSTNAME = "localhost"
        self.NEO4J_PORT = 7687
        self.NEO4J_LOGIN = "neo4j"
        self.NEO4J_PASSWORD = "password"
        self.NEO4J_DATABASE = "neo4j"
        self.RABBITMQ_ENABLED = False
        self.RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"
        self.RABBITMQ_EXCHANGE = "notes.events"
        self.RABBITMQ_NOTE_UPDATED_QUEUE = "category.note-updated"
        self.RABBITMQ_NOTE_UPDATED_ROUTING_KEY = "note.updated"

    def _read(self) -> None:
        if not os.path.exists(self._path):
            raise FileNotFoundError(f"config yaml does not exist: {self._path}")

        with open(self._path, encoding="utf-8") as config_file:
            config_yaml = yaml.safe_load(config_file.read()) or {}

        for key, value in config_yaml.items():
            field_name = key.upper()
            if field_name in self.__slots__:
                setattr(self, field_name, value)

    def _read_env(self) -> None:
        env_parsers = {
            "DEBUG": self._parse_bool,
            "LOG_LEVEL": str,
            "LOGGER_NAME": str,
            "APP_HOST": str,
            "APP_PORT": int,
            "APP_TITLE": str,
            "API_PREFIX": str,
            "CORS_ALLOWED_ORIGINS": self._parse_origins,
            "NEO4J_SCHEME": str,
            "NEO4J_HOSTNAME": str,
            "NEO4J_PORT": int,
            "NEO4J_LOGIN": str,
            "NEO4J_PASSWORD": str,
            "NEO4J_DATABASE": str,
            "RABBITMQ_ENABLED": self._parse_bool,
            "RABBITMQ_URL": str,
            "RABBITMQ_EXCHANGE": str,
            "RABBITMQ_NOTE_UPDATED_QUEUE": str,
            "RABBITMQ_NOTE_UPDATED_ROUTING_KEY": str,
        }

        for field_name, parser in env_parsers.items():
            raw_value = os.getenv(field_name)
            if raw_value is None or raw_value == "":
                continue
            setattr(self, field_name, parser(raw_value))

    @staticmethod
    def _parse_bool(value: str) -> bool:
        return value.strip().lower() in {"1", "true", "yes", "on"}

    @staticmethod
    def _parse_origins(value: str) -> list[str]:
        return [item.strip() for item in value.split(",") if item.strip()]
