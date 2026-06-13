package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"myproject/internal/apperror"
	userclient "myproject/internal/client/user"
	"myproject/internal/config"
	"myproject/pkg/logging"
	jwt2 "myproject/pkg/middleware/jwt"

	"github.com/cristalhq/jwt/v5"
)

const (
	autURL    = "/api/auth"
	signupURL = "/api/signup"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember,omitempty"`
}

type signupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Remember bool   `json:"remember,omitempty"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	Remember     bool   `json:"remember,omitempty"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type Handler struct {
	Logger      logging.Logger
	UserService userclient.UserService
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPost+" "+autURL, apperror.Middleware(h.Login))
	mux.HandleFunc(http.MethodPut+" "+autURL, apperror.Middleware(h.Refresh))
	mux.HandleFunc(http.MethodDelete+" "+autURL, apperror.Middleware(h.Logout))
	mux.HandleFunc(http.MethodPost+" "+signupURL, apperror.Middleware(h.Signup))
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) error {
	var req signupRequest
	defer r.Body.Close()

	if err := decodeJSONBody(r, &req); err != nil {
		return err
	}

	userUUID, err := h.UserService.CreateUser(r.Context(), userclient.CreateUserDTO{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	response, err := h.issueTokens(r, w, userUUID, req.Email, req.Remember)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	defer r.Body.Close()

	if err := decodeJSONBody(r, &req); err != nil {
		return err
	}

	authUser, err := h.UserService.Authenticate(r.Context(), userclient.AuthUserDTO{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	response, err := h.issueTokens(r, w, authUser.Uuid, authUser.Email, req.Remember)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	req, err := decodeOptionalRefreshRequest(r)
	if err != nil {
		return err
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return apperror.BadRequestError("refresh_token is required")
	}

	newRefreshToken := uuid.NewString()
	expiresAt := refreshExpiry().Unix()
	session, err := h.UserService.RotateSession(r.Context(), userclient.RotateUserSessionDTO{
		RefreshTokenHash:    hashRefreshToken(refreshToken),
		NewRefreshTokenHash: hashRefreshToken(newRefreshToken),
		ExpiresAt:           expiresAt,
		UserAgent:           r.UserAgent(),
		IPAddress:           clientIP(r),
	})
	if err != nil {
		normalizedErr := normalizeRefreshError(err)
		if isUnauthorizedError(normalizedErr) {
			clearRefreshTokenCookie(w)
		}
		return normalizedErr
	}

	userData, err := h.UserService.GetUser(r.Context(), session.UserUUID)
	if err != nil {
		return err
	}

	accessToken, err := buildAccessToken(session.UserUUID, userData.Email)
	if err != nil {
		h.Logger.Error("failed to build access token", "error", err)
		return apperror.APIError(http.StatusInternalServerError, "API-50001", "token generation failed", "token generation failed")
	}

	setRefreshTokenCookie(w, newRefreshToken, req.Remember)
	return writeJSON(w, http.StatusCreated, tokenResponse{Token: accessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	req, err := decodeOptionalRefreshRequest(r)
	if err != nil {
		return err
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		clearRefreshTokenCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	err = h.UserService.RevokeSession(r.Context(), hashRefreshToken(refreshToken))
	if err != nil && !isNotFoundError(err) {
		return err
	}

	clearRefreshTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) issueTokens(r *http.Request, w http.ResponseWriter, userUUID, email string, remember bool) (tokenResponse, error) {
	accessToken, err := buildAccessToken(userUUID, email)
	if err != nil {
		h.Logger.Error("failed to build access token", "error", err)
		return tokenResponse{}, apperror.APIError(http.StatusInternalServerError, "API-50001", "token generation failed", "token generation failed")
	}

	refreshToken := uuid.NewString()
	err = h.UserService.CreateSession(r.Context(), userclient.CreateUserSessionDTO{
		UserUUID:         userUUID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        refreshExpiry().Unix(),
		UserAgent:        r.UserAgent(),
		IPAddress:        clientIP(r),
	})
	if err != nil {
		return tokenResponse{}, err
	}

	setRefreshTokenCookie(w, refreshToken, remember)
	return tokenResponse{Token: accessToken}, nil
}

func buildAccessToken(userUUID, email string) (string, error) {
	cfg := config.GetConfig()
	key := []byte(cfg.JWT.Secret)

	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		return "", err
	}

	ttl := time.Duration(cfg.JWT.AccessTokenTTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	builder := jwt.NewBuilder(signer)
	claims := jwt2.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        userUUID,
			Audience:  []string{"users"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
		Email: email,
	}

	token, err := builder.Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

func refreshExpiry() time.Time {
	ttl := time.Duration(config.GetConfig().JWT.RefreshTokenTTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}

	return time.Now().Add(ttl)
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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

func decodeJSONBody(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return apperror.BadRequestError("request body is required")
		default:
			return apperror.BadRequestError("invalid JSON body")
		}
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return apperror.BadRequestError("request body must contain a single JSON object")
	}

	return nil
}

func normalizeRefreshError(err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) && appErr.StatusCode == http.StatusNotFound {
		return apperror.UnauthorizedError("invalid refresh token")
	}

	return err
}

func decodeOptionalRefreshRequest(r *http.Request) (refreshRequest, error) {
	req := refreshRequest{}

	if err := decodeOptionalJSONBody(r, &req); err != nil {
		return refreshRequest{}, err
	}

	if cookie, err := r.Cookie(config.GetConfig().RefreshCookie.Name); err == nil {
		req.RefreshToken = strings.TrimSpace(cookie.Value)
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	return req, nil
}

func decodeOptionalJSONBody(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return apperror.BadRequestError("invalid JSON body")
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return apperror.BadRequestError("request body must contain a single JSON object")
	}

	return nil
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, remember bool) {
	cfg := config.GetConfig()
	cookie := &http.Cookie{
		Name:     cfg.RefreshCookie.Name,
		Value:    refreshToken,
		Path:     cfg.RefreshCookie.Path,
		Domain:   cfg.RefreshCookie.Domain,
		HttpOnly: true,
		SameSite: parseSameSite(cfg.RefreshCookie.SameSite),
		Secure:   cfg.RefreshCookie.Secure,
	}

	if remember {
		cookie.Expires = refreshExpiry()
		cookie.MaxAge = int(time.Until(cookie.Expires).Seconds())
	}

	http.SetCookie(w, cookie)
}

func clearRefreshTokenCookie(w http.ResponseWriter) {
	cfg := config.GetConfig()
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.RefreshCookie.Name,
		Value:    "",
		Path:     cfg.RefreshCookie.Path,
		Domain:   cfg.RefreshCookie.Domain,
		HttpOnly: true,
		SameSite: parseSameSite(cfg.RefreshCookie.SameSite),
		Secure:   cfg.RefreshCookie.Secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func parseSameSite(value string) http.SameSite {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return apperror.APIError(http.StatusInternalServerError, "API-50002", "response encoding failed", "response encoding failed")
	}

	return nil
}

func isNotFoundError(err error) bool {
	var appErr *apperror.AppError
	return errors.As(err, &appErr) && appErr.StatusCode == http.StatusNotFound
}

func isUnauthorizedError(err error) bool {
	var appErr *apperror.AppError
	return errors.As(err, &appErr) && appErr.StatusCode == http.StatusUnauthorized
}
