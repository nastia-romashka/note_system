from neo4j import GraphDatabase
from config import Config
from dao.storage.storage import Storage


class Neo4jStorage(Storage):
    __slots__ = ["_database", "_driver", "url", "_username", "_password"]

    def __init__(self, config: Config) -> None:
        self.url = config.neo4j_uri
        self._username = config.NEO4J_LOGIN
        self._password = config.NEO4J_PASSWORD
        self._database = config.NEO4J_DATABASE
        self._connect()

    def _connect(self) -> None:
        self._driver = GraphDatabase.driver(
            self.url,
            auth=(self._username, self._password),
        )
        self._driver.verify_connectivity()

    def _execute_cypher(self, command: str, parameters: dict | None = None) -> list:
        with self._driver.session(database=self._database) as session:
            result = session.run(command, parameters or {})
            return [record.data() for record in result]

    def find_one(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)

    def find(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)

    def create(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)

    def update(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)

    def delete(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)

    def execute(self, cypher_cmd: str, parameters: dict | None = None) -> list:
        return self._execute_cypher(command=cypher_cmd, parameters=parameters)
