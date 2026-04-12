from abc import ABC, abstractmethod

from dao.model.dto import CreateCategoryDTO, DeleteCategoryDTO, UpdateCategoryDTO
from dao.model.model import Category


class CategoryDAO(ABC):
    @abstractmethod
    def find_user_categories(self, user_uuid: str) -> list[Category]:
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
    def create_sub_category(self, category: CreateCategoryDTO) -> Category:
        raise NotImplementedError

    @abstractmethod
    def update_category(self, category: UpdateCategoryDTO) -> None:
        raise NotImplementedError

    @abstractmethod
    def delete_category(self, category: DeleteCategoryDTO) -> None:
        raise NotImplementedError
