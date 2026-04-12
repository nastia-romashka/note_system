import logging

from dao.category.category import CategoryDAO
from dao.model.dto import CreateCategoryDTO, DeleteCategoryDTO, UpdateCategoryDTO
from dao.model.model import Category
from exceptions import AppError, NotFoundException


class CategoryService:
    __slots__ = ["category_dao", "logger"]

    def __init__(self, category_dao: CategoryDAO, logger: logging.Logger) -> None:
        self.category_dao = category_dao
        self.logger = logger

    def get_categories(self, user_uuid: str) -> list[Category]:
        is_exist = self.category_dao.check_user_exist(user_uuid=user_uuid)
        if not is_exist:
            raise NotFoundException(exc_data=AppError.USER_NOT_FOUND)

        return self.category_dao.find_user_categories(user_uuid=user_uuid)

    def create_category(self, category: CreateCategoryDTO) -> Category:
        if not category.parent_uuid or category.parent_uuid == "":
            self.logger.debug("no parent category - create root category")
            return self.category_dao.create_root_category(category=category)

        is_exist = self.category_dao.check_category_exist(
            category_uuid=category.parent_uuid
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.logger.debug("parent category is present. create sub category")
        return self.category_dao.create_sub_category(category=category)

    def update_category(self, category: UpdateCategoryDTO) -> None:
        is_exist = self.category_dao.check_category_exist(category_uuid=category.uuid)
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.update_category(category=category)

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        is_exist = self.category_dao.check_category_exist(category_uuid=category.uuid)
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.delete_category(category=category)
