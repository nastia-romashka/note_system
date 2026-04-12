from enum import Enum


class AppError(Enum):
    def __init__(self, code: str, message: str, developer_message: str) -> None:
        self.code = code
        self.message = message
        self.developer_message = developer_message

    SYSTEM_ERROR = ("CS-00001", "system error", "")
    CATEGORY_NOT_FOUND = ("CS-00008", "category not found", "")
    USER_NOT_FOUND = ("CS-00009", "user not found", "")
    VALIDATION_ERROR = ("CS-00010", "validation error", "")


class AppException(Exception):
    __slots__ = ["code", "message", "developer_message"]

    def __init__(
        self,
        exc_data: AppError | None = None,
        code: str | None = None,
        error: str | None = None,
        developer_message: str | None = None,
        *args: object,
    ) -> None:
        self.code = ""
        self.message = ""
        self.developer_message = ""

        if exc_data is not None:
            self.code = exc_data.code
            self.message = exc_data.message
            self.developer_message = exc_data.developer_message

        if code is not None:
            self.code = code
        if error is not None:
            self.message = error
        if developer_message is not None:
            self.developer_message = developer_message

        super().__init__(*args)

    def to_dict(self) -> dict[str, str]:
        return {
            "code": self.code,
            "message": self.message,
            "developer_message": self.developer_message,
        }


class NotFoundException(AppException):
    pass
