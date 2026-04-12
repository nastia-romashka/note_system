import logging
import os
from pathlib import Path

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from config import Config
from constants import EXPOSE_HEADERS, CONFIG_FILE_PATH, LOG_DIR
from di import ApplicationContainer
from exceptions import AppException
from helpers.fast import (
    app_exception_handler,
    http_exception_handler,
    uncaught_exception_handler,
    validation_exception_handler,
)
from resources.categories import router as categories_router
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException


def setup_logger(config: Config) -> logging.Logger:
    logger = logging.getLogger(config.LOGGER_NAME)
    logger.setLevel(getattr(logging, config.LOG_LEVEL.upper(), logging.DEBUG))
    logger.propagate = False

    if logger.handlers:
        return logger

    os.makedirs(LOG_DIR, exist_ok=True)

    formatter = logging.Formatter(
        "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )

    file_handler = logging.FileHandler(Path(LOG_DIR) / "all.log", encoding="utf-8")
    file_handler.setLevel(getattr(logging, config.LOG_LEVEL.upper(), logging.DEBUG))
    file_handler.setFormatter(formatter)

    stream_handler = logging.StreamHandler()
    stream_handler.setLevel(getattr(logging, config.LOG_LEVEL.upper(), logging.DEBUG))
    stream_handler.setFormatter(formatter)

    logger.addHandler(file_handler)
    logger.addHandler(stream_handler)
    return logger


class AppFactory:
    __slots__ = ["config", "logger", "container"]

    def __init__(self, config: Config, logger: logging.Logger) -> None:
        self.config = config
        self.logger = logger
        self.container = ApplicationContainer(config=config, logger=logger)

    def create_app(self) -> FastAPI:
        app = FastAPI(
            title=self.config.APP_TITLE,
            debug=self.config.DEBUG,
        )

        app.state.config = self.config
        app.state.logger = self.logger
        app.state.container = self.container

        app.add_middleware(
            CORSMiddleware,
            allow_origins=self.config.CORS_ALLOWED_ORIGINS,
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
            expose_headers=EXPOSE_HEADERS,
        )

        app.add_exception_handler(AppException, app_exception_handler)
        app.add_exception_handler(StarletteHTTPException, http_exception_handler)
        app.add_exception_handler(RequestValidationError, validation_exception_handler)
        app.add_exception_handler(Exception, uncaught_exception_handler)

        @app.get("/health")
        async def healthcheck() -> dict[str, object]:
            app.state.logger.debug("healthcheck endpoint called")
            return {
                "status": "ok",
                "service": "category_service",
                "debug": self.config.DEBUG,
                "neo4j_uri": self.config.neo4j_uri,
            }

        @app.get(f"{self.config.API_PREFIX}/config")
        async def get_config() -> dict[str, object]:
            return self.config.to_dict()

        app.include_router(categories_router)

        for route in app.routes:
            app.state.logger.debug("registered route: %s", getattr(route, "path", ""))

        return app


config_path = Path(__file__).with_name(CONFIG_FILE_PATH)
config = Config(str(config_path))
logger = setup_logger(config)
app = AppFactory(config, logger).create_app()


if __name__ == "__main__":
    uvicorn.run(
        "app:app",
        host=config.APP_HOST,
        port=config.APP_PORT,
        reload=config.DEBUG,
        log_level=config.LOG_LEVEL.lower(),
    )
