from http import HTTPStatus
from traceback import format_exc

from fastapi import Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from exceptions import AppError, AppException, NotFoundException


def app_exception_handler(_: Request, exception: AppException) -> JSONResponse:
    http_code = HTTPStatus.INTERNAL_SERVER_ERROR
    if isinstance(exception, NotFoundException):
        http_code = HTTPStatus.NOT_FOUND

    return JSONResponse(status_code=int(http_code), content=exception.to_dict())


def uncaught_exception_handler(request: Request, exception: Exception) -> JSONResponse:
    request.app.state.logger.error(
        "Uncaught exception: %s\n%s",
        exception,
        format_exc(),
    )
    return JSONResponse(
        status_code=int(HTTPStatus.INTERNAL_SERVER_ERROR),
        content=AppException(
            exc_data=AppError.SYSTEM_ERROR,
            developer_message=str(exception),
        ).to_dict(),
    )


def http_exception_handler(
    request: Request, exception: StarletteHTTPException
) -> JSONResponse:
    request.app.state.logger.warning(
        "HTTP exception: status=%s detail=%s",
        exception.status_code,
        exception.detail,
    )
    return JSONResponse(
        status_code=exception.status_code,
        content=AppException(
            code=f"CS-{exception.status_code:05d}",
            error=str(exception.detail),
            developer_message=str(exception.detail),
        ).to_dict(),
    )


def validation_exception_handler(
    request: Request, exception: RequestValidationError
) -> JSONResponse:
    request.app.state.logger.warning("Validation exception: %s", exception.errors())
    return JSONResponse(
        status_code=int(HTTPStatus.UNPROCESSABLE_ENTITY),
        content=AppException(
            exc_data=AppError.VALIDATION_ERROR,
            developer_message=str(exception.errors()),
        ).to_dict(),
    )
