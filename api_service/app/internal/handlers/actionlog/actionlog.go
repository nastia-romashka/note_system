package actionlog

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	userclient "myproject/internal/client/user"
	"myproject/pkg/logging"
)

type Recorder struct {
	Logger      logging.Logger
	UserService userclient.UserService
}

func (r Recorder) Record(req *http.Request, userUUID, action, entityType, entityID string, metadata map[string]any) {
	if r.UserService == nil {
		return
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	dto := userclient.CreateUserActionDTO{
		Action:     action,
		EntityType: entityType,
		EntityId:   entityID,
		Status:     "success",
		Metadata:   metadata,
		IPAddress:  clientIP(req),
		UserAgent:  req.UserAgent(),
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := r.UserService.CreateAction(ctx, userUUID, dto)
		if err != nil {
			r.Logger.Warn(
				"failed to record user action",
				"user_uuid", userUUID,
				"action", action,
				"entity_type", entityType,
				"entity_id", entityID,
				"error", err,
			)
		}
	}()
}

func clientIP(req *http.Request) string {
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := strings.TrimSpace(req.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(req.RemoteAddr)
}
