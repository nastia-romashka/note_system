from dao.category.category import CategoryDAO
from dao.model.dto import CreateCategoryDTO, DeleteCategoryDTO, UpdateCategoryDTO
from dao.model.model import Category
from dao.storage.storage import Storage


class Neo4jCategoryDAO(CategoryDAO):
    __slots__ = ["storage"]

    def __init__(self, storage: Storage) -> None:
        self.storage = storage

    def find_user_categories(self, user_uuid: str) -> list[Category]:
        result = self.storage.find(
            f"""
            MATCH path = (u:User)-[*]->(c)
            WHERE NOT (c)-->() AND u.id = "{user_uuid}"
            WITH collect(path) AS ps
            CALL apoc.convert.toTree(ps) YIELD value
            RETURN value
            """
        )

        if not result:
            return []

        own_categories = result[0].get("value", {}).get("own", [])
        return self._parse_categories(categories=own_categories)

    def check_user_exist(self, user_uuid: str) -> bool:
        return self._check_entity_exist(entity="User", entity_uuid=user_uuid)

    def create_root_category(self, category: CreateCategoryDTO) -> Category:
        result = self.storage.create(
            f"""
            MERGE (u:User {{id: "{category.user_uuid}"}})
            CREATE (c:Category {{name: "{category.name}", id: randomUUID()}})
            CREATE (u)-[:OWN]->(c)
            RETURN c.id AS category_id
            """
        )

        return Category(
            uuid=result[0]["category_id"],
            name=category.name,
            parent_uuid=category.parent_uuid,
            children=None,
        )

    def check_category_exist(self, category_uuid: str) -> bool:
        return self._check_entity_exist(entity="Category", entity_uuid=category_uuid)

    def create_sub_category(self, category: CreateCategoryDTO) -> Category:
        result = self.storage.create(
            f"""
            MATCH (parent:Category {{id: "{category.parent_uuid}"}})
            CREATE (c:Category {{name: "{category.name}", id: randomUUID()}})
            CREATE (parent)-[:CHILD]->(c)
            RETURN c.id AS category_id
            """
        )

        return Category(
            uuid=result[0]["category_id"],
            name=category.name,
            parent_uuid=category.parent_uuid,
            children=None,
        )

    def update_category(self, category: UpdateCategoryDTO) -> None:
        self.storage.update(
            f"""
            MATCH (c:Category {{id: "{category.uuid}"}})
            SET c.name = "{category.name}"
            """
        )

    def delete_category(self, category: DeleteCategoryDTO) -> None:
        self.storage.delete(
            f"""
            MATCH (c:Category)
            WHERE c.id = "{category.uuid}"
            OPTIONAL MATCH (c)-[*0..]->(cc:Category)
            WITH collect(DISTINCT c) + collect(DISTINCT cc) AS nodes
            UNWIND nodes AS node
            WITH node WHERE node IS NOT NULL
            DETACH DELETE node
            """
        )

    def _check_entity_exist(self, entity: str, entity_uuid: str) -> bool:
        result = self.storage.execute(
            f"""
            OPTIONAL MATCH (n:{entity} {{id: "{entity_uuid}"}})
            RETURN n IS NOT NULL AS is_exist
            """
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
                    parent_uuid=parent_uuid,
                    children=children if children else None,
                )
            )

        return parsed_categories
