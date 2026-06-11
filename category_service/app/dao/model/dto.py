from dao.model.base import Base


class CreateCategoryDTO(Base):
    workspace_id: str
    workspace_name: str | None = None
    workspace_type: str | None = None
    author_user_uuid: str
    name: str
    color: str | None = None
    parent_uuid: str | None = None


class UpdateCategoryDTO(Base):
    uuid: str | None = None
    workspace_id: str
    name: str | None = None
    color: str | None = None
    parent_uuid: str | None = None


class DeleteCategoryDTO(Base):
    uuid: str | None = None
    workspace_id: str


class CreateGraphNoteDTO(Base):
    uuid: str
    workspace_id: str
    workspace_name: str | None = None
    workspace_type: str | None = None
    author_user_uuid: str
    category_uuid: str
    header: str
    created_date: int


class UpdateGraphNoteDTO(Base):
    workspace_id: str
    category_uuid: str | None = None
    header: str | None = None


class DeleteGraphNoteDTO(Base):
    workspace_id: str


class CreateUserGraphLinkDTO(Base):
    workspace_id: str
    user_uuid: str
    source_id: str
    target_id: str


class DeleteUserGraphLinkDTO(Base):
    workspace_id: str
    user_uuid: str
    source_id: str
    target_id: str
