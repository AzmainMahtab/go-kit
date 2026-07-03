package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/application/commands"
	"github.com/elite4print/elite4print-go/internal/modules/auth/application/queries"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	identitydomain "github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/platform/http/responses"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AuthHandler holds auth HTTP handlers.
type AuthHandler struct {
	login            *commands.Login
	logout           *commands.Logout
	refresh          *commands.Refresh
	revokeSession    *commands.RevokeSession
	revokeAll        *commands.RevokeAllSessions
	listSessions     *queries.ListSessions
	v                validator.Validator
}

// NewAuthHandler creates a handler group.
func NewAuthHandler(
	login *commands.Login,
	logout *commands.Logout,
	refresh *commands.Refresh,
	revokeSession *commands.RevokeSession,
	revokeAll *commands.RevokeAllSessions,
	listSessions *queries.ListSessions,
	v validator.Validator,
) *AuthHandler {
	return &AuthHandler{
		login:         login,
		logout:        logout,
		refresh:       refresh,
		revokeSession: revokeSession,
		revokeAll:     revokeAll,
		listSessions:  listSessions,
		v:             v,
	}
}

// Login handles POST /auth/login.
//
// @Summary Authenticate a user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeAndValidate(r, h.v, &req); err != nil {
		responses.BadRequest(w, err)
		return
	}

	result, err := h.login.Handle(r.Context(), req.ToCommand(clientIP(r), r.UserAgent()))
	if err != nil {
		h.mapError(w, err)
		return
	}

	responses.OK(w, FromTokenPairResult(result))
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := CurrentUserFromContext(r.Context())
	accessJTI := r.Context().Value(accessJTIKey{}).(string)

	if err := h.logout.Handle(r.Context(), application.LogoutCommand{
		UserID:    user.UserID,
		SessionID: user.SessionID,
		AccessJTI: accessJTI,
	}); err != nil {
		h.mapError(w, err)
		return
	}

	responses.NoContent(w)
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeAndValidate(r, h.v, &req); err != nil {
		responses.BadRequest(w, err)
		return
	}

	result, err := h.refresh.Handle(r.Context(), req.ToCommand(clientIP(r), r.UserAgent()))
	if err != nil {
		h.mapError(w, err)
		return
	}

	responses.OK(w, FromTokenPairResult(result))
}

// ListSessions handles GET /auth/sessions.
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user := CurrentUserFromContext(r.Context())

	sessions, err := h.listSessions.Execute(r.Context(), application.ListSessionsQuery{UserID: user.UserID}, user.SessionID)
	if err != nil {
		h.mapError(w, err)
		return
	}

	resp := make([]SessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = FromSessionResult(s)
	}

	responses.OK(w, resp)
}

// RevokeSession handles POST /auth/sessions/{id}/revoke.
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user := CurrentUserFromContext(r.Context())
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responses.BadRequest(w, errors.New("invalid session id"))
		return
	}

	if err := h.revokeSession.Handle(r.Context(), application.RevokeSessionCommand{
		UserID:    user.UserID,
		SessionID: sessionID,
	}); err != nil {
		h.mapError(w, err)
		return
	}

	responses.NoContent(w)
}

// RevokeAllSessions handles POST /auth/sessions/revoke-all.
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := CurrentUserFromContext(r.Context())

	if err := h.revokeAll.Handle(r.Context(), application.RevokeAllSessionsCommand{
		UserID:           user.UserID,
		CurrentSessionID: user.SessionID,
	}); err != nil {
		h.mapError(w, err)
		return
	}

	responses.NoContent(w)
}

func (h *AuthHandler) mapError(w http.ResponseWriter, err error) {
	var valErr *validator.ValidationError

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		responses.Unauthorized(w, "invalid email or password")
	case errors.Is(err, domain.ErrSessionNotFound), errors.Is(err, domain.ErrSessionExpired), errors.Is(err, domain.ErrSessionRevoked):
		responses.Unauthorized(w, err.Error())
	case errors.Is(err, domain.ErrTokenBlacklisted):
		responses.Unauthorized(w, "token revoked")
	case errors.Is(err, identitydomain.ErrUserInactive):
		responses.Unauthorized(w, "account inactive")
	case errors.Is(err, identitydomain.ErrUserSuspended):
		responses.Forbidden(w, "account suspended")
	case errors.As(err, &valErr):
		responses.BadRequest(w, err)
	default:
		responses.InternalError(w)
	}
}

func clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func decodeAndValidate(r *http.Request, v validator.Validator, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return v.ValidateStruct(dst)
}
