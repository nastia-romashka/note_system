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

    def get_categories(self, workspace_id: str) -> list[Category]:
        return self.category_dao.find_workspace_categories(workspace_id=workspace_id)

    def get_stats(self, workspace_id: str) -> CategoryStats:
        return self.category_dao.count_workspace_categories(workspace_id=workspace_id)

    def get_workspace_graph(self, workspace_id: str) -> GraphData:
        return self.category_dao.find_workspace_graph(workspace_id=workspace_id)

    def create_category(self, category: CreateCategoryDTO) -> Category:
        if not category.parent_uuid or category.parent_uuid == "":
            self.logger.debug("no parent category - create root category")
            return self.category_dao.create_root_category(category=category)

        is_exist = self.category_dao.check_category_in_workspace(
            category_uuid=category.parent_uuid,
            workspace_id=category.workspace_id,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.logger.debug("parent category is present. create sub category")
        return self.category_dao.create_sub_category(category=category)

    def update_category(self, category: UpdateCategoryDTO) -> None:
        is_exist = self._category_exists_for_request(
            category_uuid=category.uuid,
            workspace_id=category.workspace_id,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.update_category(category=category)

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        is_exist = self._category_exists_for_request(
            category_uuid=category.uuid,
            workspace_id=category.workspace_id,
        )
        if not is_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.delete_category(category=category)

    def create_note_node(self, note: CreateGraphNoteDTO) -> None:
        is_category_exist = self.category_dao.check_category_in_workspace(
            category_uuid=note.category_uuid,
            workspace_id=note.workspace_id,
        )
        if not is_category_exist:
            raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.create_note(note=note)

    def update_note_node(self, note_uuid: str, note: UpdateGraphNoteDTO) -> None:
        is_note_exist = self.category_dao.check_note_in_workspace(
            note_uuid=note_uuid,
            workspace_id=note.workspace_id,
        )
        if not is_note_exist:
            raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)

        if note.category_uuid:
            is_category_exist = self.category_dao.check_category_in_workspace(
                category_uuid=note.category_uuid,
                workspace_id=note.workspace_id,
            )
            if not is_category_exist:
                raise NotFoundException(exc_data=AppError.CATEGORY_NOT_FOUND)

        self.category_dao.update_note(note_uuid=note_uuid, note=note)

    def delete_note_node(self, note_uuid: str, workspace_id: str) -> None:
        is_note_exist = self.category_dao.check_note_in_workspace(
            note_uuid=note_uuid,
            workspace_id=workspace_id,
        )
        if not is_note_exist:
            raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)

        self.category_dao.delete_note(note_uuid=note_uuid, workspace_id=workspace_id)

    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        self._ensure_graph_entity_in_workspace(link.source_id, link.workspace_id)
        self._ensure_graph_entity_in_workspace(link.target_id, link.workspace_id)

        self.category_dao.create_user_graph_link(link=link)

    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        self._ensure_graph_entity_in_workspace(link.source_id, link.workspace_id)
        self._ensure_graph_entity_in_workspace(link.target_id, link.workspace_id)

        self.category_dao.delete_user_graph_link(link=link)

    def ensure_workspace_member(
        self,
        workspace_id: str,
        user_uuid: str,
        workspace_name: str | None = None,
        workspace_type: str | None = None,
    ) -> None:
        self.category_dao.ensure_workspace_member(
            workspace_id=workspace_id,
            user_uuid=user_uuid,
            workspace_name=workspace_name,
            workspace_type=workspace_type,
        )

    def _category_exists_for_request(
        self,
        category_uuid: str | None,
        workspace_id: str | None,
    ) -> bool:
        if not category_uuid or not workspace_id:
            return False

        return self.category_dao.check_category_in_workspace(
            category_uuid=category_uuid,
            workspace_id=workspace_id,
        )

    def _ensure_graph_entity_in_workspace(self, entity_id: str, workspace_id: str) -> None:
        if self.category_dao.check_category_in_workspace(entity_id, workspace_id):
            return
        if self.category_dao.check_note_in_workspace(entity_id, workspace_id):
            return

        raise NotFoundException(exc_data=AppError.NOTE_NOT_FOUND)
