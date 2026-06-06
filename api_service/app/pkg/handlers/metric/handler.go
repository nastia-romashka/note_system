package metric

import (
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
	"net/http"
)

const (
	URL = "/api/heartbeat"
)

type Handler struct {
	Logger logging.Logger
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+URL, jwt.JWTMiddleware(h.Heartbeat))
}

func (h *Handler) Heartbeat(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}
