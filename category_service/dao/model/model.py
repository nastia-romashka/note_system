from __future__ import annotations

from dao.model.base import Base


class Category(Base):
    uuid: str
    name: str
    parent_uuid: str | None = None
    children: list["Category"] | None = None


Category.model_rebuild()
