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


class CreateGraphNoteDTO(Base):
    uuid: str
    user_uuid: str
    category_uuid: str
    header: str


class UpdateGraphNoteDTO(Base):
    user_uuid: str
    category_uuid: str | None = None
    header: str | None = None


class DeleteGraphNoteDTO(Base):
    user_uuid: str


class LinkGraphNotesDTO(Base):
    user_uuid: str
