package controllers

import (
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type UsersController struct {
	GetUserProfile *usecases.GetUserProfile
}

type UserProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Handle returns the current authenticated user profile
// @Summary			Get Current User Profile
// @Description Retrieves detailed information about the currently logged-in user. Requires a valid Bearer Token.
// @Security		BearerAuth
// @Tags				Users
// @Produce 		json
// @Success			200	{object}	controllers.UserProfileResponse "User profile details"
// @Failure			401	{object}	common.ErrResponse	"Unauthorized / Invalid Token"
// @Failure			404	{object}	common.ErrResponse	"User not found"
// @Failure     500  {object}  common.ErrResponse "Internal server error"
// @Router			/users/me [get]
func (uc *UsersController) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated in context",
		})
		return
	}

	user, err := uc.GetUserProfile.Exec(r.Context(), userID)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
			Message: err.Error(),
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		AvatarURL: user.AvatarUrl,
		CreatedAt: user.CreatedAt,
	})
}
