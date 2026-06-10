import logging

from dao.category.category import CategoryDAO
from dao.model.dto import (
    CreateCategoryDTO,
    CreateUserGraphLinkDTO,
    CreateGraphNoteDTO,
    DeleteCategoryDTO,
    DeleteUserGraphLinkDTO,
    UpdateCategoryDTO,
    UpdateGraphNoteDTO,
)
from dao.model.model import Category, CategoryStats, GraphData
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

    def get_stats(self, user_uuid: str) -> CategoryStats:
        return self.category_dao.count_user_categories(user_uuid=user_uuid)

    def get_user_graph(self, user_uuid: str) -> GraphData:
        return self.category_dao.find_user_graph(user_uuid=user_uuid)

    def create_category(self, category: CreateCategoryDTO) -> Category:
        if not category.parent_uuid or category.parent_uuid == "":
            self.logger.debug("no parent category - create root category")
            return self.category_dao.create_root_category(category=category)

        is_exist = self.category_dao.check_category_belongs_to_user(
            category_uuid=category.parent_uuid,
            user_uuid=category.user_uuid,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.logger.debug("parent category is present. create sub category")
        return self.category_dao.create_sub_category(category=category)

    def update_category(self, category: UpdateCategoryDTO) -> None:
        is_exist = self._category_exists_for_request(
            category_uuid=category.uuid,
            user_uuid=category.user_uuid,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.update_category(category=category)

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        is_exist = self._category_exists_for_request(
            category_uuid=category.uuid,
            user_uuid=category.user_uuid,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.delete_category(category=category)

    def create_note_node(self, note: CreateGraphNoteDTO) -> None:
        is_category_exist = self.category_dao.check_category_belongs_to_user(
            category_uuid=note.category_uuid,
            user_uuid=note.user_uuid,
        )
        if not is_category_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.create_note(note=note)

    def update_note_node(self, note_uuid: str, note: UpdateGraphNoteDTO) -> None:
        is_note_exist = self.category_dao.check_note_belongs_to_user(
            note_uuid=note_uuid,
            user_uuid=note.user_uuid,
        )
        if not is_note_exist:
            raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)

        if note.category_uuid:
            is_category_exist = self.category_dao.check_category_belongs_to_user(
                category_uuid=note.category_uuid,
                user_uuid=note.user_uuid,
            )
            if not is_category_exist:
                raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.update_note(note_uuid=note_uuid, note=note)

    def delete_note_node(self, note_uuid: str, user_uuid: str) -> None:
        is_note_exist = self.category_dao.check_note_belongs_to_user(
            note_uuid=note_uuid,
            user_uuid=user_uuid,
        )
        if not is_note_exist:
            raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)

        self.category_dao.delete_note(note_uuid=note_uuid, user_uuid=user_uuid)

    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        self._ensure_graph_entity_belongs_to_user(link.source_id, link.user_uuid)
        self._ensure_graph_entity_belongs_to_user(link.target_id, link.user_uuid)

        self.category_dao.create_user_graph_link(link=link)

    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        self._ensure_graph_entity_belongs_to_user(link.source_id, link.user_uuid)
        self._ensure_graph_entity_belongs_to_user(link.target_id, link.user_uuid)

        self.category_dao.delete_user_graph_link(link=link)

    def _category_exists_for_request(
        self,
        category_uuid: str | None,
        user_uuid: str | None,
    ) -> bool:
        if not category_uuid:
            return False

        if user_uuid:
            return self.category_dao.check_category_belongs_to_user(
                category_uuid=category_uuid,
                user_uuid=user_uuid,
            )

        return self.category_dao.check_category_exist(category_uuid=category_uuid)

    def _ensure_graph_entity_belongs_to_user(self, entity_id: str, user_uuid: str) -> None:
        if self.category_dao.check_category_belongs_to_user(entity_id, user_uuid):
            return
        if self.category_dao.check_note_belongs_to_user(entity_id, user_uuid):
            return

        raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)
