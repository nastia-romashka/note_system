from __future__ import annotations

from dao.model.base import Base


class Category(Base):
    uuid: str
    name: str
    user_uuid: str | None = None
    color: str | None = None
    created_at: int | None = None
    parent_uuid: str | None = None
    children: list["Category"] | None = None


class CategoryStats(Base):
    categories_count: int


class GraphNote(Base):
    uuid: str
    user_uuid: str
    category_uuid: str
    header: str
    created_date: int | None = None


class GraphNode(Base):
    id: str
    type: str
    label: str
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
