from dao.category.category import CategoryDAO
from dao.model.dto import (
    CreateCategoryDTO,
    CreateUserGraphLinkDTO,
    CreateGraphNoteDTO,
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

    def find_user_categories(self, user_uuid: str) -> list[Category]:
        result = self.storage.find(
            """
            MATCH path = (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(c:Category)
            WHERE NOT (c)-[:CHILD]->(:Category)
            WITH collect(path) AS paths
            CALL apoc.paths.toJsonTree(paths) YIELD value
            RETURN value
            """,
            {"user_uuid": user_uuid},
        )

        if not result:
            return []

        own_categories = result[0].get("value", {}).get("own", [])
        return self._parse_categories(categories=own_categories)

    def count_user_categories(self, user_uuid: str) -> CategoryStats:
        result = self.storage.execute(
            """
            MATCH (u:User {id: $user_uuid})
            OPTIONAL MATCH (u)-[:OWN|CHILD*1..]->(c:Category)
            RETURN count(DISTINCT c) AS categories_count
            """,
            {"user_uuid": user_uuid},
        )

        if not result:
            return CategoryStats(categories_count=0)

        return CategoryStats(categories_count=result[0]["categories_count"])

    def find_user_graph(self, user_uuid: str) -> GraphData:
        category_rows = self.storage.find(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(category:Category)
            RETURN DISTINCT
                category.id AS id,
                category.name AS label,
                category.color AS color,
                category.created_at AS created_at
            """,
            {"user_uuid": user_uuid},
        )

        if not category_rows:
            return GraphData(nodes=[], edges=[])

        category_nodes = [
            GraphNode(
                id=row["id"],
                type="category",
                label=row["label"],
                color=row.get("color"),
                created_at=row.get("created_at"),
            )
            for row in category_rows
        ]
        category_ids = [node.id for node in category_nodes]

        category_edge_rows = self.storage.find(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(parent:Category)-[:CHILD]->(child:Category)
            RETURN DISTINCT
                parent.id AS source,
                child.id AS target,
                "CHILD" AS type
            """,
            {"user_uuid": user_uuid},
        )

        note_rows = self.storage.find(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(category:Category)-[:HAS_NOTE]->(note:Note {user_uuid: $user_uuid})
            RETURN DISTINCT
                category.id AS category_id,
                note.uuid AS id,
                note.header AS label,
                note.category_uuid AS category_uuid,
                note.created_date AS created_at
            """,
            {"user_uuid": user_uuid},
        )

        note_nodes = [
            GraphNode(
                id=row["id"],
                type="note",
                label=row["label"],
                category_uuid=row.get("category_uuid"),
                created_at=row.get("created_at"),
            )
            for row in note_rows
        ]
        note_ids = [node.id for node in note_nodes]

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
            CALL {
                WITH $category_ids AS category_ids
                UNWIND category_ids AS category_id
                MATCH (source:Category {id: category_id})-[custom:USER_LINK]->(target)
                RETURN source, custom, target
                UNION
                WITH $note_ids AS note_ids
                UNWIND note_ids AS note_id
                MATCH (source:Note {uuid: note_id})-[custom:USER_LINK]->(target)
                RETURN source, custom, target
            }
            WITH source, custom, target
            WHERE
                (target:Category AND target.id IN $category_ids)
                OR
                (target:Note AND target.uuid IN $note_ids)
            RETURN DISTINCT
                coalesce(source.id, source.uuid) AS source,
                coalesce(target.id, target.uuid) AS target,
                type(custom) AS type
            """,
            {
                "category_ids": category_ids,
                "note_ids": note_ids,
            },
        )

        nodes = category_nodes + note_nodes
        edge_rows = (
            category_edge_rows
            + has_note_edge_rows
            + custom_graph_edge_rows
        )
        edges = [
            GraphEdge(**edge)
            for edge in edge_rows
            if edge.get("source") and edge.get("target") and edge.get("type")
        ]
        return GraphData(nodes=nodes, edges=edges)

    def check_user_exist(self, user_uuid: str) -> bool:
        return self._check_entity_exist(
            entity="User",
            property_name="id",
            property_value=user_uuid,
        )

    def create_root_category(self, category: CreateCategoryDTO) -> Category:
        color = category.color or DEFAULT_CATEGORY_COLOR
        result = self.storage.create(
            """
            MERGE (u:User {id: $user_uuid})
            CREATE (c:Category {
                name: $name,
                id: randomUUID(),
                user_uuid: $user_uuid,
                color: $color,
                created_at: toInteger(timestamp() / 1000)
            })
            CREATE (u)-[:OWN]->(c)
            RETURN c.id AS category_id, c.created_at AS category_created_at
            """,
            {
                "user_uuid": category.user_uuid,
                "name": category.name,
                "color": color,
            },
        )

        return Category(
            uuid=result[0]["category_id"],
            name=category.name,
            user_uuid=category.user_uuid,
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

    def check_category_belongs_to_user(self, category_uuid: str, user_uuid: str) -> bool:
        result = self.storage.execute(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(c:Category {id: $category_uuid})
            RETURN c IS NOT NULL AS is_exist
            """,
            {
                "user_uuid": user_uuid,
                "category_uuid": category_uuid,
            },
        )
        return bool(result and result[0]["is_exist"])

    def check_note_belongs_to_user(self, note_uuid: str, user_uuid: str) -> bool:
        result = self.storage.execute(
            """
            MATCH (note:Note {uuid: $note_uuid, user_uuid: $user_uuid})
            RETURN note IS NOT NULL AS is_exist
            """,
            {
                "note_uuid": note_uuid,
                "user_uuid": user_uuid,
            },
        )
        return bool(result and result[0]["is_exist"])

    def create_sub_category(self, category: CreateCategoryDTO) -> Category:
        color = category.color or DEFAULT_CATEGORY_COLOR
        result = self.storage.create(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(parent:Category {id: $parent_uuid})
            CREATE (c:Category {
                name: $name,
                id: randomUUID(),
                user_uuid: $user_uuid,
                color: $color,
                created_at: toInteger(timestamp() / 1000)
            })
            CREATE (parent)-[:CHILD]->(c)
            RETURN c.id AS category_id, c.created_at AS category_created_at
            """,
            {
                "user_uuid": category.user_uuid,
                "parent_uuid": category.parent_uuid,
                "name": category.name,
                "color": color,
            },
        )

        return Category(
            uuid=result[0]["category_id"],
            name=category.name,
            user_uuid=category.user_uuid,
            color=color,
            created_at=result[0].get("category_created_at"),
            parent_uuid=category.parent_uuid,
            children=None,
        )

    def update_category(self, category: UpdateCategoryDTO) -> None:
        self.storage.update(
            """
            MATCH (c:Category {id: $category_uuid})
            WHERE $user_uuid IS NULL OR c.user_uuid = $user_uuid
            SET
                c.name = coalesce($name, c.name),
                c.color = coalesce($color, c.color)
            """,
            {
                "category_uuid": category.uuid,
                "user_uuid": category.user_uuid,
                "name": category.name,
                "color": category.color,
            },
        )

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        self.storage.delete(
            """
            MATCH (c:Category {id: $category_uuid})
            WHERE $user_uuid IS NULL OR c.user_uuid = $user_uuid
            OPTIONAL MATCH (c)-[:CHILD*0..]->(child:Category)
            WITH collect(DISTINCT c) + collect(DISTINCT child) AS raw_categories
            UNWIND raw_categories AS category_node
            WITH collect(DISTINCT category_node) AS categories
            UNWIND categories AS category_node
            OPTIONAL MATCH (category_node)-[:HAS_NOTE]->(note:Note)
            WITH categories, collect(DISTINCT note) AS notes
            WITH categories + notes AS nodes
            UNWIND nodes AS node
            WITH DISTINCT node
            WHERE node IS NOT NULL
            DETACH DELETE node
            """,
            {
                "category_uuid": category.uuid,
                "user_uuid": category.user_uuid,
            },
        )

    def create_note(self, note: CreateGraphNoteDTO) -> None:
        self.storage.create(
            """
            MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(category:Category {id: $category_uuid})
            MERGE (note:Note {uuid: $note_uuid})
            ON CREATE SET
                note.user_uuid = $user_uuid,
                note.created_date = $created_date
            WITH category, note
            WHERE note.user_uuid = $user_uuid
            SET
                note.category_uuid = $category_uuid,
                note.header = $header,
                note.created_date = coalesce(note.created_date, $created_date)
            MERGE (category)-[:HAS_NOTE]->(note)
            WITH category, note
            MATCH (oldCategory:Category)-[oldRelation:HAS_NOTE]->(note)
            WHERE oldCategory.id <> category.id
            DELETE oldRelation
            """,
            {
                "note_uuid": note.uuid,
                "user_uuid": note.user_uuid,
                "category_uuid": note.category_uuid,
                "header": note.header,
                "created_date": note.created_date,
            },
        )

    def update_note(self, note_uuid: str, note: UpdateGraphNoteDTO) -> None:
        if note.category_uuid:
            self.storage.update(
                """
                MATCH (u:User {id: $user_uuid})-[:OWN|CHILD*1..]->(category:Category {id: $category_uuid})
                MATCH (note:Note {uuid: $note_uuid, user_uuid: $user_uuid})
                OPTIONAL MATCH (:Category)-[oldRelation:HAS_NOTE]->(note)
                WITH category, note, [relation IN collect(oldRelation) WHERE relation IS NOT NULL] AS oldRelations
                FOREACH (relation IN oldRelations | DELETE relation)
                SET
                    note.category_uuid = $category_uuid,
                    note.header = coalesce($header, note.header)
                MERGE (category)-[:HAS_NOTE]->(note)
                """,
                {
                    "note_uuid": note_uuid,
                    "user_uuid": note.user_uuid,
                    "category_uuid": note.category_uuid,
                    "header": note.header,
                },
            )
            return

        self.storage.update(
            """
            MATCH (note:Note {uuid: $note_uuid, user_uuid: $user_uuid})
            SET note.header = coalesce($header, note.header)
            """,
            {
                "note_uuid": note_uuid,
                "user_uuid": note.user_uuid,
                "header": note.header,
            },
        )

    def delete_note(self, note_uuid: str, user_uuid: str) -> None:
        self.storage.delete(
            """
            MATCH (note:Note {uuid: $note_uuid, user_uuid: $user_uuid})
            DETACH DELETE note
            """,
            {
                "note_uuid": note_uuid,
                "user_uuid": user_uuid,
            },
        )

    def create_user_graph_link(self, link: CreateUserGraphLinkDTO) -> None:
        self.storage.create(
            """
            MATCH (source)
            WHERE
                (
                    source:Category AND source.id = $source_id AND source.user_uuid = $user_uuid
                )
                OR
                (
                    source:Note AND source.uuid = $source_id AND source.user_uuid = $user_uuid
                )
            MATCH (target)
            WHERE
                (
                    target:Category AND target.id = $target_id AND target.user_uuid = $user_uuid
                )
                OR
                (
                    target:Note AND target.uuid = $target_id AND target.user_uuid = $user_uuid
                )
            WITH source, target
            WHERE elementId(source) <> elementId(target)
            MERGE (source)-[:USER_LINK]->(target)
            """,
            {
                "source_id": link.source_id,
                "target_id": link.target_id,
                "user_uuid": link.user_uuid,
            },
        )

    def delete_user_graph_link(self, link: DeleteUserGraphLinkDTO) -> None:
        self.storage.delete(
            """
            MATCH (source)-[relation:USER_LINK]->(target)
            WHERE
                (
                    source:Category AND source.id = $source_id AND source.user_uuid = $user_uuid
                    OR
                    source:Note AND source.uuid = $source_id AND source.user_uuid = $user_uuid
                )
                AND
                (
                    target:Category AND target.id = $target_id AND target.user_uuid = $user_uuid
                    OR
                    target:Note AND target.uuid = $target_id AND target.user_uuid = $user_uuid
                )
            DELETE relation
            """,
            {
                "source_id": link.source_id,
                "target_id": link.target_id,
                "user_uuid": link.user_uuid,
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
            OPTIONAL MATCH (n:{entity} {{{property_name}: $property_value}})
            RETURN n IS NOT NULL AS is_exist
            """,
            {"property_value": property_value},
        )
        return bool(result and result[0]["is_exist"])

    def _parse_categories(
        self, categories: list[dict], parent_uuid: str | None = None
    ) -> list[Category]:
        parsed_categories: list[Category] = []

        for category_data in categories:
            child_items = category_data.get("child", [])
            children = self._parse_categories(
                categories=child_items,
                parent_uuid=category_data["id"],
            )
            parsed_categories.append(
                Category(
                    uuid=category_data["id"],
                    name=category_data["name"],
                    user_uuid=category_data.get("user_uuid"),
                    color=category_data.get("color"),
                    created_at=category_data.get("created_at"),
                    parent_uuid=parent_uuid,
                    children=children if children else None,
                )
            )

        return parsed_categories

    def _prepare_graph_schema(self) -> None:
        commands = [
            "CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE",
            "CREATE CONSTRAINT category_id_unique IF NOT EXISTS FOR (c:Category) REQUIRE c.id IS UNIQUE",
            "CREATE CONSTRAINT note_uuid_unique IF NOT EXISTS FOR (n:Note) REQUIRE n.uuid IS UNIQUE",
            """
            MATCH (u:User)-[:OWN|CHILD*1..]->(c:Category)
            WHERE c.user_uuid IS NULL
            SET c.user_uuid = u.id
            """,
            """
            MATCH (c:Category)
            WHERE c.color IS NULL
            SET c.color = $default_color
            """,
        ]

        for command in commands:
            try:
                self.storage.execute(command, {"default_color": DEFAULT_CATEGORY_COLOR})
            except Exception:
                # Existing duplicate data should not prevent the service from starting.
                # The graph can still be cleaned manually before constraints are retried.
                continue
