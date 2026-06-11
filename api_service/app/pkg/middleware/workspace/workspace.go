package workspace

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"myproject/internal/apperror"
	userclient "myproject/internal/client/user"
	"myproject/internal/requestctx"
)

const headerWorkspaceID = "X-Workspace-Id"

func Middleware(userService userclient.UserService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUUID, err := requestctx.UserUUID(r)
		if err != nil {
			next(w, r)
			return
		}

		workspaceID := resolveWorkspaceID(r)
		if workspaceID == "" {
			workspace, err := userService.GetPersonalWorkspace(r.Context(), userUUID)
			if err != nil {
				writeError(w, err)
				return
			}
			workspaceID = workspace.Uuid
		}

		access, err := userService.GetWorkspaceAccess(r.Context(), workspaceID, userUUID)
		if err != nil {
			writeError(w, err)
			return
		}

		workspaceType := "shared"
		if access.Workspace.IsPersonal {
			workspaceType = "personal"
		}

		ctx := context.WithValue(r.Context(), "workspace_id", access.Workspace.Uuid)
		ctx = context.WithValue(ctx, "workspace_role", access.Role)
		ctx = context.WithValue(ctx, "workspace_visibility", access.Workspace.Visibility)
		ctx = context.WithValue(ctx, "workspace_name", access.Workspace.Name)
		ctx = context.WithValue(ctx, "workspace_type", workspaceType)
		next(w, r.WithContext(ctx))
	}
}

func resolveWorkspaceID(r *http.Request) string {
	if value := strings.TrimSpace(r.URL.Query().Get("workspace_id")); value != "" {
		return value
	}

	if value := strings.TrimSpace(r.Header.Get(headerWorkspaceID)); value != "" {
		return value
	}

	if value := workspaceIDFromPath(r.URL.Path); value != "" {
		return value
	}

	return ""
}

func workspaceIDFromPath(requestPath string) string {
	trimmed := strings.Trim(strings.TrimSpace(requestPath), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 {
		return ""
	}

	if parts[0] != "api" || parts[1] != "workspaces" {
		return ""
	}

	if parts[2] == "" || parts[2] == "invites" {
		return ""
	}

	return parts[2]
}

func writeError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":              appErr.ErrorCode,
		"message":           appErr.Message,
		"developer_message": appErr.DeveloperMessage,
	})
}
