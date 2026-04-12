from py2neo import Graph
from config import Config
from dao.storage.storage import Storage


class Neo4jStorage(Storage):
    __slots__ = ["_graph", "url", "_username", "_password"]

    def __init__(self, config: Config) -> None:
        self.url = config.neo4j_uri
        self._username = config.NEO4J_LOGIN
        self._password = config.NEO4J_PASSWORD
        self._connect()

    def _connect(self) -> None:
        self._graph = Graph(self.url, user=self._username, password=self._password)

    def _execute_cypher(self, command: str) -> list:
        cursor = self._graph.run(command)
        return cursor.data()

    def find_one(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)

    def find(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)

    def create(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)

    def update(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)

    def delete(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)

    def execute(self, cypher_cmd: str) -> list:
        return self._execute_cypher(command=cypher_cmd)
