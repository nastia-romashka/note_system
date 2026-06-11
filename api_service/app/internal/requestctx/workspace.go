package requestctx

import (
	"net/http"

	"myproject/internal/apperror"
)

func WorkspaceID(r *http.Request) (string, error) {
	rawWorkspaceID := r.Context().Value("workspace_id")
	if rawWorkspaceID == nil {
		return "", apperror.BadRequestError("there is no workspace_id in context")
	}

	workspaceID, ok := rawWorkspaceID.(string)
	if !ok || workspaceID == "" {
		return "", apperror.BadRequestError("there is no workspace_id in context")
	}

	return workspaceID, nil
}

func WorkspaceRole(r *http.Request) (string, error) {
	rawWorkspaceRole := r.Context().Value("workspace_role")
	if rawWorkspaceRole == nil {
		return "", apperror.BadRequestError("there is no workspace_role in context")
	}

	workspaceRole, ok := rawWorkspaceRole.(string)
	if !ok || workspaceRole == "" {
		return "", apperror.BadRequestError("there is no workspace_role in context")
	}

	return workspaceRole, nil
}

func WorkspaceName(r *http.Request) string {
	rawWorkspaceName := r.Context().Value("workspace_name")
	workspaceName, _ := rawWorkspaceName.(string)
	return workspaceName
}

func WorkspaceType(r *http.Request) string {
	rawWorkspaceType := r.Context().Value("workspace_type")
	workspaceType, _ := rawWorkspaceType.(string)
	return workspaceType
}
