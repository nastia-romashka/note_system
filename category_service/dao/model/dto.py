from dao.model.base import Base


class CreateCategoryDTO(Base):
    user_uuid: str
    name: str
    color: str | None = None
    parent_uuid: str | None = None


class UpdateCategoryDTO(Base):
    uuid: str | None = None
    name: str
    color: str | None = None
    parent_uuid: str | None = None


class DeleteCategoryDTO(Base):
    uuid: str | None = None
    user_uuid: str | None = None
