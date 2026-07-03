package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/application/commands"
	"github.com/elite4print/elite4print-go/internal/modules/identity/application/queries"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/platform/http/responses"
	"github.com/elite4print/elite4print-go/internal/shared/apperrors"
	"github.com/elite4print/elite4print-go/internal/shared/pagination"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// UserHandler holds the identity HTTP handlers.
type UserHandler struct {
	register  *commands.RegisterUser
	update    *commands.UpdateUser
	getUser   *queries.GetUser
	listUsers *queries.ListUsers
	v         validator.Validator
}

// NewUserHandler creates a handler group.
func NewUserHandler(
	register *commands.RegisterUser,
	update *commands.UpdateUser,
	getUser *queries.GetUser,
	listUsers *queries.ListUsers,
	v validator.Validator,
) *UserHandler {
	return &UserHandler{
		register:  register,
		update:    update,
		getUser:   getUser,
		listUsers: listUsers,
		v:         v,
	}
}

// Register handles POST /users/register.
//
// @Summary Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param body body RegisterUserRequest true "User registration details"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /users/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := decodeAndValidate(r, h.v, &req); err != nil {
		responses.BadRequest(w, err)
		return
	}

	result, err := h.register.Handle(r.Context(), req.ToCommand())
	if err != nil {
		h.mapError(w, err)
		return
	}

	responses.Created(w, FromUserResult(result))
}

// GetByID handles GET /users/{id}.
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responses.BadRequest(w, errors.New("invalid user id"))
		return
	}

	result, err := h.getUser.ByID(r.Context(), application.GetUserByIDQuery{UserID: id})
	if err != nil {
		h.mapError(w, err)
		return
	}

	responses.OK(w, FromUserResult(result))
}

// Update handles PATCH /users/{id}.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		responses.BadRequest(w, errors.New("invalid user id"))
		return
	}

	var req UpdateUserRequest
	if err := decodeAndValidate(r, h.v, &req); err != nil {
		responses.BadRequest(w, err)
		return
	}

	cmd := req.ToCommand()
	cmd.UserID = id

	result, err := h.update.Handle(r.Context(), cmd)
	if err != nil {
		h.mapError(w, err)
		return
	}

	responses.OK(w, FromUserResult(result))
}

// List handles GET /users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = pagination.DefaultPageSize
	}

	page, err := h.listUsers.Execute(r.Context(), application.ListUsersQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		h.mapError(w, err)
		return
	}

	items := make([]UserResponse, len(page.Items))
	for i, item := range page.Items {
		items[i] = FromUserResult(item)
	}

	responses.JSON(w, http.StatusOK, responses.Success(http.StatusOK, ListUsersResponse{
		Items:      items,
		Total:      page.Total,
		Offset:     page.Offset,
		Limit:      page.Limit,
		TotalPages: page.TotalPages(),
	}))
}

func (h *UserHandler) mapError(w http.ResponseWriter, err error) {
	var valErr *validator.ValidationError
	var appErr *apperrors.AppError

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		responses.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		responses.Conflict(w, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrEmailRequired):
		responses.BadRequest(w, err)
	case errors.Is(err, domain.ErrWeakPassword), errors.Is(err, domain.ErrPasswordMismatch):
		responses.BadRequest(w, err)
	case errors.As(err, &valErr):
		responses.BadRequest(w, err)
	case errors.As(err, &appErr):
		responses.JSON(w, appErr.StatusCode, responses.FromAppError(appErr))
	default:
		responses.InternalError(w)
	}
}

// decodeAndValidate decodes JSON and validates a request DTO.
func decodeAndValidate(r *http.Request, v validator.Validator, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return v.ValidateStruct(dst)
}
