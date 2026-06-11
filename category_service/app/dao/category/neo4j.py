from dao.category.category import CategoryDAO
from dao.model.dto import (
    CreateCategoryDTO,
    CreateGraphNoteDTO,
    CreateUserGraphLinkDTO,
    DeleteCategoryDTO,
    DeleteUserGraphLinkDTO,
    UpdateCategoryDTO,
    UpdateGraphNoteDTO,
)
from dao.model.model import Category, CategoryStats, GraphData, GraphEdge, GraphNode
from dao.storage.storage import Storage


DEFAULT_CATEGORY_COLOR = "#8FA3FF"


class Neo4jCategoryDAO(CategoryDAO):
    __slots__ = ["storage"]

    def __init__(self, storage: Storage) -> None:
        self.storage = storage
        self._prepare_graph_schema()

    def find_workspace_categories(self, workspace_id: str) -> list[Category]:
        rows = self.storage.find(
            """
            MATCH (workspace:Workspace {id: $workspace_id})
            OPTIONAL MATCH (workspace)-[:HAS_CATEGORY|CHILD*1..]->(category:Category)
            WITH DISTINCT category
            WHERE category IS NOT NULL
            OPTIONAL MATCH (parent:Category {workspace_id: $workspace_id})-[:CHILD]->(category)
            RETURN
                category.id AS id,
                category.workspace_id AS workspace_id,
                category.author_user_uuid AS author_user_uuid,
                category.name AS name,
                category.color AS color,
                category.created_at AS created_at,
                parent.id AS parent_uuid
            ORDER BY category.created_at ASC, category.name ASC
            """,
            {"workspace_id": workspace_id},
        )

        return self._build_category_tree(rows)

    def count_workspace_categories(self, workspace_id: str) -> CategoryStats:
        result = self.storage.execute(
            """
            MATCH (workspace:Workspace {id: $workspace_id})
            OPTIONAL MATCH (workspace)-[:HAS_CATEGORY|CHILD*1..]->(category:Category)
            RETURN count(DISTINCT category) AS categories_count
            """,
            {"workspace_id": workspace_id},
        )

        if not result:
            return CategoryStats(categories_count=0)

        return CategoryStats(categories_count=result[0]["categories_count"])

    def find_workspace_graph(self, workspace_id: str) -> GraphData:
        category_rows = self.storage.find(
            """
            MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(category:Category)
            RETURN DISTINCT
                category.id AS id,
                category.workspace_id AS workspace_id,
                category.author_user_uuid AS author_user_uuid,
                category.name AS label,
                category.color AS color,
                category.created_at AS created_at
            """,
            {"workspace_id": workspace_id},
        )

        if not category_rows:
            return GraphData(nodes=[], edges=[])

        category_nodes = [
            GraphNode(
                id=row["id"],
                type="category",
                label=row["label"],
                workspace_id=row.get("workspace_id"),
                author_user_uuid=row.get("author_user_uuid"),
                color=row.get("color"),
                created_at=row.get("created_at"),
            )
            for row in category_rows
        ]

        category_edge_rows = self.storage.find(
            """
            MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(parent:Category)-[:CHILD]->(child:Category)
            RETURN DISTINCT
                parent.id AS source,
                child.id AS target,
                "CHILD" AS type
            """,
            {"workspace_id": workspace_id},
        )

        note_rows = self.storage.find(
            """
            MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(category:Category)-[:HAS_NOTE]->(note:Note {workspace_id: $workspace_id})
            RETURN DISTINCT
                category.id AS category_id,
                note.uuid AS id,
                note.workspace_id AS workspace_id,
                note.author_user_uuid AS author_user_uuid,
                note.header AS label,
                note.category_uuid AS category_uuid,
                note.created_at AS created_at
            """,
            {"workspace_id": workspace_id},
        )

        note_nodes = [
            GraphNode(
                id=row["id"],
                type="note",
                label=row["label"],
                workspace_id=row.get("workspace_id"),
                author_user_uuid=row.get("author_user_uuid"),
                category_uuid=row.get("category_uuid"),
                created_at=row.get("created_at"),
            )
            for row in note_rows
        ]

        has_note_edge_rows = [
            {
                "source": row["category_id"],
                "target": row["id"],
                "type": "HAS_NOTE",
            }
            for row in note_rows
            if row.get("category_id") and row.get("id")
        ]

        custom_graph_edge_rows = self.storage.find(
            """
            MATCH (source)-[relation:USER_LINK {workspace_id: $workspace_id}]->(target)
            WHERE
                (
                    source:Category AND source.workspace_id = $workspace_id
                    OR
                    source:Note AND source.workspace_id = $workspace_id
                )
                AND
                (
                    target:Category AND target.workspace_id = $workspace_id
                    OR
                    target:Note AND target.workspace_id = $workspace_id
                )
            RETURN DISTINCT
                coalesce(source.id, source.uuid) AS source,
                coalesce(target.id, target.uuid) AS target,
                type(relation) AS type
            """,
            {"workspace_id": workspace_id},
        )

        nodes = category_nodes + note_nodes
        edge_rows = category_edge_rows + has_note_edge_rows + custom_graph_edge_rows
        edges = [
            GraphEdge(**edge)
            for edge in edge_rows
            if edge.get("source") and edge.get("target") and edge.get("type")
        ]

        return GraphData(nodes=nodes, edges=edges)

    def create_root_category(self, category: CreateCategoryDTO) -> Category:
        color = category.color or DEFAULT_CATEGORY_COLOR
        result = self.storage.create(
            """
            MERGE (workspace:Workspace {id: $workspace_id})
            SET
                workspace.name = coalesce($workspace_name, workspace.name),
                workspace.type = coalesce($workspace_type, workspace.type)
            MERGE (user:User {id: $author_user_uuid})
            MERGE (user)-[:MEMBER_OF]->(workspace)
            CREATE (category:Category {
                id: randomUUID(),
                workspace_id: $workspace_id,
                author_user_uuid: $author_user_uuid,
                name: $name,
                color: $color,
                created_at: toInteger(timestamp() / 1000)
            })
            CREATE (workspace)-[:HAS_CATEGORY]->(category)
            RETURN category.id AS category_id, category.created_at AS category_created_at
            """,
            {
                "workspace_id": category.workspace_id,
                "workspace_name": category.workspace_name,
                "workspace_type": category.workspace_type,
                "author_user_uuid": category.author_user_uuid,
                "name": category.name,
                "color": color,
            },
        )

        return Category(
            uuid=result[0]["category_id"],
            workspace_id=category.workspace_id,
            author_user_uuid=category.author_user_uuid,
            name=category.name,
            color=color,
            created_at=result[0].get("category_created_at"),
            parent_uuid=category.parent_uuid,
            children=None,
        )

    def check_category_exist(self, category_uuid: str) -> bool:
        return self._check_entity_exist(
            entity="Category",
            property_name="id",
            property_value=category_uuid,
        )

    def check_category_in_workspace(self, category_uuid: str, workspace_id: str) -> bool:
        result = self.storage.execute(
            """
            MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(category:Category {id: $category_uuid})
            RETURN category IS NOT NULL AS is_exist
            """,
            {
                "workspace_id": workspace_id,
                "category_uuid": category_uuid,
            },
        )
        return bool(result and result[0]["is_exist"])

    def check_note_in_workspace(self, note_uuid: str, workspace_id: str) -> bool:
        result = self.storage.execute(
            """
            MATCH (note:Note {uuid: $note_uuid, workspace_id: $workspace_id})
            RETURN note IS NOT NULL AS is_exist
            """,
            {
                "note_uuid": note_uuid,
                "workspace_id": workspace_id,
            },
        )
        return bool(result and result[0]["is_exist"])

    def create_sub_category(self, category: CreateCategoryDTO) -> Category:
        color = category.color or DEFAULT_CATEGORY_COLOR
        result = self.storage.create(
            """
            MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(parent:Category {id: $parent_uuid})
            SET
                workspace.name = coalesce($workspace_name, workspace.name),
                workspace.type = coalesce($workspace_type, workspace.type)
            MERGE (user:User {id: $author_user_uuid})
            MERGE (user)-[:MEMBER_OF]->(workspace)
            CREATE (category:Category {
                id: randomUUID(),
                workspace_id: $workspace_id,
                author_user_uuid: $author_user_uuid,
                name: $name,
                color: $color,
                created_at: toInteger(timestamp() / 1000)
            })
            CREATE (parent)-[:CHILD]->(category)
            RETURN category.id AS category_id, category.created_at AS category_created_at
            """,
            {
                "workspace_id": category.workspace_id,
                "workspace_name": category.workspace_name,
                "workspace_type": category.workspace_type,
                "author_user_uuid": category.author_user_uuid,
                "parent_uuid": category.parent_uuid,
                "name": category.name,
                "color": color,
            },
        )

        return Category(
            uuid=result[0]["category_id"],
            workspace_id=category.workspace_id,
            author_user_uuid=category.author_user_uuid,
            name=category.name,
            color=color,
            created_at=result[0].get("category_created_at"),
            parent_uuid=category.parent_uuid,
            children=None,
        )

    def update_category(self, category: UpdateCategoryDTO) -> None:
        self.storage.update(
            """
            MATCH (category:Category {id: $category_uuid, workspace_id: $workspace_id})
            SET
                category.name = coalesce($name, category.name),
                category.color = coalesce($color, category.color)
            """,
            {
                "category_uuid": category.uuid,
                "workspace_id": category.workspace_id,
                "name": category.name,
                "color": category.color,
            },
        )

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        self.storage.delete(
            """
            MATCH (category:Category {id: $category_uuid, workspace_id: $workspace_id})
            OPTIONAL MATCH (category)-[:CHILD*0..]->(child:Category {workspace_id: $workspace_id})
            WITH collect(DISTINCT category) + collect(DISTINCT child) AS raw_categories
            UNWIND raw_categories AS category_node
            WITH collect(DISTINCT category_node) AS categories
            UNWIND categories AS category_node
            OPTIONAL MATCH (category_node)-[:HAS_NOTE]->(note:Note {workspace_id: $workspace_id})
            WITH categories, collect(DISTINCT note) AS notes
            WITH categories + notes AS nodes
            UNWIND nodes AS node
            WITH DISTINCT node
            WHERE node IS NOT NULL
            DETACH DELETE node
            """,
            {
                "category_uuid": category.uuid,
                "workspace_id": category.workspace_id,
            },
        )

    def create_note(self, note: CreateGraphNoteDTO) -> None:
        self.storage.create(
            """
            MERGE (workspace:Workspace {id: $workspace_id})
            SET
                workspace.name = coalesce($workspace_name, workspace.name),
                workspace.type = coalesce($workspace_type, workspace.type)
            WITH workspace
            MATCH (workspace)-[:HAS_CATEGORY|CHILD*1..]->(category:Category {id: $category_uuid})
            MERGE (user:User {id: $author_user_uuid})
            MERGE (user)-[:MEMBER_OF]->(workspace)
            MERGE (note:Note {uuid: $note_uuid})
            ON CREATE SET
                note.workspace_id = $workspace_id,
                note.author_user_uuid = $author_user_uuid,
                note.created_at = $created_at
            WITH category, note
            WHERE note.workspace_id = $workspace_id
            SET
                note.category_uuid = $category_uuid,
                note.header = $header,
                note.created_at = coalesce(note.created_at, $created_at),
                note.author_user_uuid = coalesce(note.author_user_uuid, $author_user_uuid)
            MERGE (category)-[:HAS_NOTE]->(note)
            WITH category, note
            MATCH (old_category:Category)-[old_relation:HAS_NOTE]->(note)
            WHERE old_category.id <> category.id
            DELETE old_relation
            """,
            {
                "note_uuid": note.uuid,
                "workspace_id": note.workspace_id,
                "workspace_name": note.workspace_name,
                "workspace_type": note.workspace_type,
                "author_user_uuid": note.author_user_uuid,
                "category_uuid": note.category_uuid,
                "header": note.header,
                "created_at": note.created_date,
            },
        )

    def update_note(self, note_uuid: str, note: UpdateGraphNoteDTO) -> None:
        if note.category_uuid:
            self.storage.update(
                """
                MATCH (workspace:Workspace {id: $workspace_id})-[:HAS_CATEGORY|CHILD*1..]->(category:Category {id: $category_uuid})
                MATCH (note:Note {uuid: $note_uuid, workspace_id: $workspace_id})
                OPTIONAL MATCH (:Category {workspace_id: $workspace_id})-[old_relation:HAS_NOTE]->(note)
                WITH category, note, [relation IN collect(old_relation) WHERE relation IS NOT NULL] AS old_relations
                FOREACH (relation IN old_relations | DELETE relation)
                SET
                    note.category_uuid = $category_uuid,
                    note.header = coalesce($header, note.header)
                MERGE (category)-[:HAS_NOTE]->(note)
                """,
                {
                    "note_uuid": note_uuid,
                    "workspace_id": note.workspace_id,
                    "category_uuid": note.category_uuid,
                    "header": note.header,
                },
            )
            return

        self.storage.update(
            """
            MATCH (note:Note {uuid: $note_uuid, workspace_id: $workspace_id})
            SET note.header = coalesce($header, note.header)
            """,
            {
                "note_uuid": note_uuid,
                "workspace_id": note.workspace_id,
                "header": note.header,
            },
        )

    def delete_note(self, note_uuid: str, workspace_id: str) -> None:
        self.storage.delete(
            """
            MATCH (note:Note {uuid: $note_uuid, workspace_id: $workspace_id})
            DETACH DELETE note
            """,
            {
                "note_uuid": note_uuid,
                "workspace_id": workspace_id,
            },
        )

    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        self.storage.create(
            """
            MATCH (source)
            WHERE
                (
                    source:Category AND source.id = $source_id AND source.workspace_id = $workspace_id
                )
                OR
                (
                    source:Note AND source.uuid = $source_id AND source.workspace_id = $workspace_id
                )
            MATCH (target)
            WHERE
                (
                    target:Category AND target.id = $target_id AND target.workspace_id = $workspace_id
                )
                OR
                (
                    target:Note AND target.uuid = $target_id AND target.workspace_id = $workspace_id
                )
            WITH source, target
            WHERE elementId(source) <> elementId(target)
            MERGE (source)-[relation:USER_LINK {workspace_id: $workspace_id}]->(target)
            ON CREATE SET relation.user_uuid = $user_uuid
            """,
            {
                "workspace_id": link.workspace_id,
                "source_id": link.source_id,
                "target_id": link.target_id,
                "user_uuid": link.user_uuid,
            },
        )

    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        self.storage.delete(
            """
            MATCH (source)-[relation:USER_LINK {workspace_id: $workspace_id}]->(target)
            WHERE
                (
                    source:Category AND source.id = $source_id AND source.workspace_id = $workspace_id
                    OR
                    source:Note AND source.uuid = $source_id AND source.workspace_id = $workspace_id
                )
                AND
                (
                    target:Category AND target.id = $target_id AND target.workspace_id = $workspace_id
                    OR
                    target:Note AND target.uuid = $target_id AND target.workspace_id = $workspace_id
                )
            DELETE relation
            """,
            {
                "workspace_id": link.workspace_id,
                "source_id": link.source_id,
                "target_id": link.target_id,
            },
        )

    def ensure_workspace_member(
        self,
        workspace_id: str,
        user_uuid: str,
        workspace_name: str | None = None,
        workspace_type: str | None = None,
    ) -> None:
        self.storage.create(
            """
            MERGE (workspace:Workspace {id: $workspace_id})
            SET
                workspace.name = coalesce($workspace_name, workspace.name),
                workspace.type = coalesce($workspace_type, workspace.type)
            MERGE (user:User {id: $user_uuid})
            MERGE (user)-[:MEMBER_OF]->(workspace)
            """,
            {
                "workspace_id": workspace_id,
                "workspace_name": workspace_name,
                "workspace_type": workspace_type,
                "user_uuid": user_uuid,
            },
        )

    def _check_entity_exist(
        self,
        entity: str,
        property_name: str,
        property_value: str,
    ) -> bool:
        result = self.storage.execute(
            f"""
            OPTIONAL MATCH (node:{entity} {{{property_name}: $property_value}})
            RETURN node IS NOT NULL AS is_exist
            """,
            {"property_value": property_value},
        )
        return bool(result and result[0]["is_exist"])

    def _build_category_tree(self, rows: list[dict]) -> list[Category]:
        if not rows:
            return []

        categories_by_id: dict[str, Category] = {}
        parent_by_id: dict[str, str | None] = {}

        for row in rows:
            category_id = row.get("id")
            if not category_id:
                continue

            categories_by_id[category_id] = Category(
                uuid=category_id,
                workspace_id=row["workspace_id"],
                author_user_uuid=row.get("author_user_uuid"),
                name=row["name"],
                color=row.get("color"),
                created_at=row.get("created_at"),
                parent_uuid=row.get("parent_uuid"),
                children=[],
            )
            parent_by_id[category_id] = row.get("parent_uuid")

        roots: list[Category] = []
        for category_id, category in categories_by_id.items():
            parent_uuid = parent_by_id.get(category_id)
            if not parent_uuid or parent_uuid not in categories_by_id:
                roots.append(category)
                continue

            parent = categories_by_id[parent_uuid]
            if parent.children is None:
                parent.children = []
            parent.children.append(category)

        self._normalize_children(roots)
        return sorted(roots, key=self._category_sort_key)

    def _normalize_children(self, categories: list[Category]) -> None:
        for category in categories:
            if category.children:
                category.children = sorted(category.children, key=self._category_sort_key)
                self._normalize_children(category.children)
            else:
                category.children = None

    def _category_sort_key(self, category: Category) -> tuple[int, str]:
        created_at = category.created_at or 0
        return (created_at, category.name.lower())

    def _prepare_graph_schema(self) -> None:
        commands = [
            "CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (user:User) REQUIRE user.id IS UNIQUE",
            "CREATE CONSTRAINT workspace_id_unique IF NOT EXISTS FOR (workspace:Workspace) REQUIRE workspace.id IS UNIQUE",
            "CREATE CONSTRAINT category_id_unique IF NOT EXISTS FOR (category:Category) REQUIRE category.id IS UNIQUE",
            "CREATE CONSTRAINT note_uuid_unique IF NOT EXISTS FOR (note:Note) REQUIRE note.uuid IS UNIQUE",
            """
            MATCH (category:Category)
            WHERE category.color IS NULL
            SET category.color = $default_color
            """,
        ]

        for command in commands:
            try:
                self.storage.execute(command, {"default_color": DEFAULT_CATEGORY_COLOR})
            except Exception:
                # Existing duplicate data should not prevent the service from starting.
                continue
