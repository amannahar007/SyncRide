package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/syncride/backend/auth-service/internal/application/usecase"
	"github.com/syncride/backend/shared/constants"
	sharedErr "github.com/syncride/backend/shared/errors"
	"github.com/syncride/backend/shared/utils"
)

type AuthHandler struct {
	authUC   usecase.AuthUseCase
	validate *validator.Validate
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUC:   authUC,
		validate: validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req usecase.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("invalid request payload"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("validation failed: "+err.Error()))
		return
	}

	resp, err := h.authUC.Register(r.Context(), &req)
	if err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp, "user registered successfully")
}

// Login authenticates a user
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecase.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("invalid request payload"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("validation failed: "+err.Error()))
		return
	}

	resp, err := h.authUC.Login(r.Context(), &req)
	if err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp, "login successful")
}

// RefreshToken handles token rotation
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req usecase.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("invalid request payload"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("validation failed: "+err.Error()))
		return
	}

	resp, err := h.authUC.RefreshToken(r.Context(), &req)
	if err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp, "token refreshed successfully")
}

// Logout revokes the provided refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req usecase.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("invalid request payload"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewBadRequestError("validation failed: "+err.Error()))
		return
	}

	err := h.authUC.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil, "logout successful")
}

// GetProfile retrieves the currently authenticated user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Retrieve UserID from context (injected by JWT middleware)
	userIDStr, ok := r.Context().Value(constants.CtxKeyUserID).(string)
	if !ok || userIDStr == "" {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewUnauthorizedError("missing user identity in context"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewUnauthorizedError("invalid user ID format"))
		return
	}

	user, err := h.authUC.GetProfile(r.Context(), userID)
	if err != nil {
		// Use error unwrapping to return specific HTTP error correctly
		var appErr *sharedErr.AppError
		if errors.As(err, &appErr) {
			sharedErr.WriteHTTPError(w, r.URL.Path, appErr)
		} else {
			sharedErr.WriteHTTPError(w, r.URL.Path, sharedErr.NewInternalError("failed to retrieve profile", err))
		}
		return
	}

	// Prepare user DTO
	type UserDTO struct {
		ID        string   `json:"id"`
		Email     string   `json:"email"`
		Roles     []string `json:"roles"`
		Status    string   `json:"status"`
		CreatedAt string   `json:"created_at"`
	}

	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	dto := UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Roles:     roles,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.WriteJSON(w, http.StatusOK, dto, "profile retrieved successfully")
}
