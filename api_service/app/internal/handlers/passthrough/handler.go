package passthrough

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"myproject/internal/apperror"
	"myproject/internal/proxy"
	"myproject/internal/requestctx"
	"myproject/pkg/logging"
	"myproject/pkg/middleware/jwt"
)

const (
	categoriesURL  = "/api/categories"
	graphURL       = "/api/graph"
	graphLinksURL  = "/api/graph/links"
	calendarURL    = "/api/calendar"
	noteURL        = "/api/notes/{uuid}"
	tagsURL        = "/api/tags"
	searchNotesURL = "/api/search/notes"
	meURL          = "/api/me"
	meActionsURL   = "/api/me/actions"
)

type Handler struct {
	Logger          logging.Logger
	CategoryService proxy.Service
	NoteService     proxy.Service
	UserService     proxy.Service
	SearchService   proxy.Service
}

func NewHandler(
	logger logging.Logger,
	categoryBaseURL string,
	noteBaseURL string,
	userBaseURL string,
	searchBaseURL string,
) Handler {
	return Handler{
		Logger:          logger,
		CategoryService: proxy.NewService(categoryBaseURL, 20*time.Second),
		NoteService:     proxy.NewService(noteBaseURL, 20*time.Second),
		UserService:     proxy.NewService(userBaseURL, 20*time.Second),
		SearchService:   proxy.NewService(searchBaseURL, 20*time.Second),
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+categoriesURL, jwt.JWTMiddleware(apperror.Middleware(h.GetCategories)))
	mux.HandleFunc(http.MethodGet+" "+graphURL, jwt.JWTMiddleware(apperror.Middleware(h.GetGraph)))
	mux.HandleFunc(http.MethodPost+" "+graphLinksURL, jwt.JWTMiddleware(apperror.Middleware(h.CreateGraphLink)))
	mux.HandleFunc(http.MethodDelete+" "+graphLinksURL, jwt.JWTMiddleware(apperror.Middleware(h.DeleteGraphLink)))
	mux.HandleFunc(http.MethodGet+" "+calendarURL, jwt.JWTMiddleware(apperror.Middleware(h.GetCalendarNotes)))
	mux.HandleFunc(http.MethodGet+" "+noteURL, jwt.JWTMiddleware(apperror.Middleware(h.GetNote)))
	mux.HandleFunc(http.MethodGet+" "+tagsURL, jwt.JWTMiddleware(apperror.Middleware(h.GetTags)))
	mux.HandleFunc(http.MethodGet+" "+searchNotesURL, jwt.JWTMiddleware(apperror.Middleware(h.SearchNotes)))
	mux.HandleFunc(http.MethodGet+" "+meURL, jwt.JWTMiddleware(apperror.Middleware(h.GetProfile)))
	mux.HandleFunc(http.MethodGet+" "+meActionsURL, jwt.JWTMiddleware(apperror.Middleware(h.GetActions)))
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.CategoryService.Forward(
		w,
		r,
		http.MethodGet,
		"categories",
		withUserUUID(r.URL.Query(), userUUID),
		nil,
		nil,
	)
}

func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.CategoryService.Forward(
		w,
		r,
		http.MethodGet,
		"graph",
		withUserUUID(r.URL.Query(), userUUID),
		nil,
		nil,
	)
}

func (h *Handler) CreateGraphLink(w http.ResponseWriter, r *http.Request) error {
	return h.forwardGraphLink(w, r, http.MethodPost)
}

func (h *Handler) DeleteGraphLink(w http.ResponseWriter, r *http.Request) error {
	return h.forwardGraphLink(w, r, http.MethodDelete)
}

func (h *Handler) GetCalendarNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	query := withUserUUID(r.URL.Query(), userUUID)
	if query.Get("from") == "" {
		return apperror.BadRequestError("from is required")
	}
	if query.Get("to") == "" {
		return apperror.BadRequestError("to is required")
	}

	return h.NoteService.Forward(w, r, http.MethodGet, "calendar", query, nil, nil)
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) error {
	noteUUID := r.PathValue("uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("empty note uuid")
	}

	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.NoteService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("notes/%s", noteUUID),
		withUserUUID(r.URL.Query(), userUUID),
		nil,
		nil,
	)
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.NoteService.Forward(
		w,
		r,
		http.MethodGet,
		"tags",
		withUserUUID(r.URL.Query(), userUUID),
		nil,
		nil,
	)
}

func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.SearchService.Forward(
		w,
		r,
		http.MethodGet,
		"search/notes",
		withUserUUID(r.URL.Query(), userUUID),
		nil,
		nil,
	)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/profile", userUUID),
		nil,
		nil,
		nil,
	)
}

func (h *Handler) GetActions(w http.ResponseWriter, r *http.Request) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	limit, err := positiveIntQuery(r, "limit", 50)
	if err != nil {
		return err
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := positiveIntQuery(r, "offset", 0)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	return h.UserService.Forward(
		w,
		r,
		http.MethodGet,
		fmt.Sprintf("users/%s/actions", userUUID),
		query,
		nil,
		nil,
	)
}

func (h *Handler) forwardGraphLink(w http.ResponseWriter, r *http.Request, method string) error {
	userUUID, err := requestctx.UserUUID(r)
	if err != nil {
		return err
	}

	var payload map[string]any
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return apperror.BadRequestError("can't decode")
	}

	sourceID := strings.TrimSpace(asString(payload["source_id"]))
	targetID := strings.TrimSpace(asString(payload["target_id"]))
	if sourceID == "" || targetID == "" {
		return apperror.BadRequestError("source_id and target_id are required")
	}

	payload["source_id"] = sourceID
	payload["target_id"] = targetID
	payload["user_uuid"] = userUUID

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal graph link payload: %w", err)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	return h.CategoryService.Forward(
		w,
		r,
		method,
		"graph/links",
		nil,
		bytes.NewReader(data),
		headers,
	)
}

func withUserUUID(query url.Values, userUUID string) url.Values {
	cloned := cloneQuery(query)
	cloned.Set("user_uuid", userUUID)
	return cloned
}

func cloneQuery(query url.Values) url.Values {
	cloned := make(url.Values, len(query))
	for key, values := range query {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}

func positiveIntQuery(r *http.Request, name string, defaultValue int) (int, error) {
	rawValue := r.URL.Query().Get(name)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil || value < 0 {
		return 0, apperror.BadRequestError(fmt.Sprintf("invalid %s", name))
	}

	return value, nil
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}
