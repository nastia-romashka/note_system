from __future__ import annotations

from dao.model.base import Base


class Category(Base):
    uuid: str
    workspace_id: str
    author_user_uuid: str | None = None
    name: str
    color: str | None = None
    created_at: int | None = None
    parent_uuid: str | None = None
    children: list["Category"] | None = None


class CategoryStats(Base):
    categories_count: int


class GraphNote(Base):
    uuid: str
    workspace_id: str
    author_user_uuid: str | None = None
    category_uuid: str
    header: str
    created_at: int | None = None


class GraphNode(Base):
    id: str
    type: str
    label: str
    workspace_id: str | None = None
    author_user_uuid: str | None = None
    color: str | None = None
    category_uuid: str | None = None
    created_at: int | None = None


class GraphEdge(Base):
    source: str
    target: str
    type: str


class GraphData(Base):
    nodes: list[GraphNode]
    edges: list[GraphEdge]


Category.model_rebuild()
