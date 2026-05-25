from logging import Logger

from dependency_injector import containers, providers

from config import Config
from dao.category.neo4j import Neo4jCategoryDAO
from dao.storage.neo4j import Neo4jStorage
from service import CategoryService


class StorageModule(containers.DeclarativeContainer):
    config = providers.Dependency(instance_of=Config)

    storage = providers.Singleton(Neo4jStorage, config=config)
    category_dao = providers.Singleton(Neo4jCategoryDAO, storage=storage)


class LoggerModule(containers.DeclarativeContainer):
    logger = providers.Dependency(instance_of=Logger)


class ApplicationContainer(containers.DeclarativeContainer):
    config = providers.Dependency(instance_of=Config)
    logger = providers.Dependency(instance_of=Logger)

    storage_module = providers.Container(StorageModule, config=config)
    logger_module = providers.Container(LoggerModule, logger=logger)

    category_service = providers.Singleton(
        CategoryService,
        category_dao=storage_module.category_dao,
        logger=logger_module.logger,
    )
