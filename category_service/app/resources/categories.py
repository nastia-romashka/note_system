from http import HTTPStatus

from fastapi import APIRouter, Body, Depends, Query, Request, Response, status

from constants import LOCATION
from dao.model.dto import (
    CreateCategoryDTO,
    CreateUserGraphLinkDTO,
    CreateGraphNoteDTO,
    DeleteCategoryDTO,
    DeleteGraphNoteDTO,
    DeleteUserGraphLinkDTO,
    UpdateCategoryDTO,
    UpdateGraphNoteDTO,
)
from dao.model.model import Category, CategoryStats, GraphData
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
    workspace_id: str = Query(...),
    service: CategoryService = Depends(get_category_service),
) -> list[Category]:
    request.app.state.logger.debug("get categories for workspace_id=%s", workspace_id)
    return service.get_categories(workspace_id=workspace_id)


@router.get(
    "/api/stats",
    response_model=CategoryStats,
    status_code=status.HTTP_200_OK,
)
async def get_stats(
    request: Request,
    workspace_id: str = Query(...),
    service: CategoryService = Depends(get_category_service),
) -> CategoryStats:
    request.app.state.logger.debug("get category stats for workspace_id=%s", workspace_id)
    return service.get_stats(workspace_id=workspace_id)


@router.get(
    "/api/graph",
    response_model=GraphData,
    status_code=status.HTTP_200_OK,
)
async def get_graph(
    request: Request,
    workspace_id: str = Query(...),
    service: CategoryService = Depends(get_category_service),
) -> GraphData:
    request.app.state.logger.debug("get graph for workspace_id=%s", workspace_id)
    return service.get_workspace_graph(workspace_id=workspace_id)


@router.post(
    "/api/graph/notes",
    status_code=status.HTTP_201_CREATED,
)
async def create_note_node(
    request: Request,
    note_dto: CreateGraphNoteDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    service.create_note_node(note=note_dto)
    request.app.state.logger.debug("created graph note uuid=%s", note_dto.uuid)
    return Response(status_code=HTTPStatus.CREATED)


@router.patch(
    "/api/graph/notes/{note_uuid}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def update_note_node(
    request: Request,
    note_uuid: str,
    note_dto: UpdateGraphNoteDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    service.update_note_node(note_uuid=note_uuid, note=note_dto)
    request.app.state.logger.debug("updated graph note uuid=%s", note_uuid)
    return Response(status_code=HTTPStatus.NO_CONTENT)


@router.delete(
    "/api/graph/notes/{note_uuid}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_note_node(
    request: Request,
    note_uuid: str,
    note_dto: DeleteGraphNoteDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    service.delete_note_node(note_uuid=note_uuid, workspace_id=note_dto.workspace_id)
    request.app.state.logger.debug("deleted graph note uuid=%s", note_uuid)
    return Response(status_code=HTTPStatus.NO_CONTENT)


@router.post(
    "/api/graph/links",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def create_user_graph_link(
    request: Request,
    link_dto: CreateUserGraphLinkDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    service.create_user_graph_link(link=link_dto)
    request.app.state.logger.debug(
        "created user graph link source=%s target=%s",
        link_dto.source_id,
        link_dto.target_id,
    )
    return Response(status_code=HTTPStatus.NO_CONTENT)


@router.delete(
    "/api/graph/links",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_user_graph_link(
    request: Request,
    link_dto: DeleteUserGraphLinkDTO,
    service: CategoryService = Depends(get_category_service),
) -> Response:
    service.delete_user_graph_link(link=link_dto)
    request.app.state.logger.debug(
        "deleted user graph link source=%s target=%s",
        link_dto.source_id,
        link_dto.target_id,
    )
    return Response(status_code=HTTPStatus.NO_CONTENT)


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
    request.app.state.category_events_publisher.publish_category_updated(
        category_uuid=cuuid,
        workspace_id=category_dto.workspace_id,
        actor_user_uuid=category_dto.actor_user_uuid,
        name=category_dto.name,
    )
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
