from abc import ABC, abstractmethod

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


class CategoryDAO(ABC):
    @abstractmethod
    def find_workspace_categories(self, workspace_id: str) -> list[Category]:
        raise NotImplementedError

    @abstractmethod
    def count_workspace_categories(self, workspace_id: str) -> CategoryStats:
        raise NotImplementedError

    @abstractmethod
    def find_workspace_graph(self, workspace_id: str) -> GraphData:
        raise NotImplementedError

    @abstractmethod
    def create_root_category(self, category: CreateCategoryDTO) -> Category:
        raise NotImplementedError

    @abstractmethod
    def check_category_exist(self, category_uuid: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def check_category_in_workspace(self, category_uuid: str, workspace_id: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def check_note_in_workspace(self, note_uuid: str, workspace_id: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def create_sub_category(self, category: CreateCategoryDTO) -> Category:
        raise NotImplementedError

    @abstractmethod
    def update_category(self, category: UpdateCategoryDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def delete_category(self, category: DeleteCategoryDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def create_note(self, note: CreateGraphNoteDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def update_note(self, note_uuid: str, note: UpdateGraphNoteDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def delete_note(self, note_uuid: str, workspace_id: str) -> None:
        raise NotImplementedError

    @abstractmethod
    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def ensure_workspace_member(
        self,
        workspace_id: str,
        user_uuid: str,
        workspace_name: str | None = None,
        workspace_type: str | None = None,
    ) -> None:
        raise NotImplementedError
