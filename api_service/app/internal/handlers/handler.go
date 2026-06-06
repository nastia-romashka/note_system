package handlers

import "net/http"

type Handler interface {
	Register(mux *http.ServeMux)
}
