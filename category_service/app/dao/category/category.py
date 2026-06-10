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
    def find_user_categories(self, user_uuid: str) -> list[Category]:
        raise NotImplementedError

    @abstractmethod
    def count_user_categories(self, user_uuid: str) -> CategoryStats:
        raise NotImplementedError

    @abstractmethod
    def find_user_graph(self, user_uuid: str) -> GraphData:
        raise NotImplementedError

    @abstractmethod
    def check_user_exist(self, user_uuid: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def create_root_category(self, category: CreateCategoryDTO) -> Category:
        raise NotImplementedError

    @abstractmethod
    def check_category_exist(self, category_uuid: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def check_category_belongs_to_user(self, category_uuid: str, user_uuid: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    def check_note_belongs_to_user(self, note_uuid: str, user_uuid: str) -> bool:
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
    def delete_note(self, note_uuid: str, user_uuid: str) -> None:
        raise NotImplementedError

    @abstractmethod
    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        raise NotImplementedError
