from http import HTTPStatus

from fastapi import APIRouter, Body, Depends, Query, Request, Response, status

from constants import LOCATION
from dao.model.dto import CreateCategoryDTO, DeleteCategoryDTO, UpdateCategoryDTO
from dao.model.model import Category
from service import CategoryService

router = APIRouter(tags=["categories"])


def get_category_service(request: Request) -> CategoryService:
    return request.app.state.container.category_service()


@router.get(
    "/api/categories",
    response_model=list[Category],
    status_code=status.HTTP_200_OK,
)
async def get_categories(
    request: Request,
    user_uuid: str = Query(...),
    service: CategoryService = Depends(get_category_service),
) -> list[Category]:
    request.app.state.logger.debug("get categories for user_uuid=%s", user_uuid)
    return service.get_categories(user_uuid=user_uuid)


@router.post(
    "/api/categories",
    status_code=status.HTTP_201_CREATED,
)
async def create_category(
    request: Request,
    category_dto: CreateCategoryDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    category = service.create_category(category=category_dto)
    request.app.state.logger.debug("created category uuid=%s", category.uuid)
    return Response(
        status_code=HTTPStatus.CREATED,
        headers={LOCATION: f"/api/categories/{category.uuid}"},
    )


@router.patch(
    "/api/categories/{cuuid}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def update_category(
    request: Request,
    cuuid: str,
    category_dto: UpdateCategoryDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    category_dto = category_dto.model_copy(update={"uuid": cuuid})
    service.update_category(category=category_dto)
    request.app.state.logger.debug("updated category uuid=%s", cuuid)
    return Response(status_code=HTTPStatus.NO_CONTENT)


@router.delete(
    "/api/categories/{cuuid}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_category(
    request: Request,
    cuuid: str,
    category_dto: DeleteCategoryDTO | None = Body(default=None),
    service: CategoryService = Depends(get_category_service),
) -> Response:
    delete_dto = (category_dto or DeleteCategoryDTO()).model_copy(update={"uuid": cuuid})
    service.delete_category(category=delete_dto)
    request.app.state.logger.debug("deleted category uuid=%s", cuuid)
    return Response(status_code=HTTPStatus.NO_CONTENT)
