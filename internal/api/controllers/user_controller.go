package controllers

import (
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type UsersController struct {
	GetUserProfile usecases.GetUserProfileUC
}

type UserProfileResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Email       string               `json:"email"`
	AvatarUrl   *string              `json:"avatarUrl"`
	Residences  []ResidenceResponse  `json:"residences"`
	Memberships []MembershipResponse `json:"memberships"`
}

type ResidenceResponse struct {
	CondominiumName string    `json:"condominiumName"`
	CondominiumID   uuid.UUID `json:"condominiumId"`
	Block           *string   `json:"block"`
	Number          string    `json:"number"`
	Type            string    `json:"type"`          // owner, tenant
	IsResponsible   bool      `json:"isResponsible"` // Pode abrir chamados?
}

type MembershipResponse struct {
	CondominiumName string    `json:"condominiumName"`
	CondominiumID   uuid.UUID `json:"condominiumId"`
	Role            string    `json:"role"` // admin, syndic
}

// Handle returns the current authenticated user profile
// @Summary			Get Current User Profile
// @Description Retrieves detailed information about the currently logged-in user. Requires a valid Bearer Token.
// @Security		BearerAuth
// @Tags				Users
// @Produce 		json
// @Success			200	{object}	controllers.UserProfileResponse   "User profile details"
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

	result, err := uc.GetUserProfile.Exec(r.Context(), userID)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
			Message: err.Error(),
		})
		return
	}

	response := UserProfileResponse{
		ID:          result.User.ID,
		Name:        result.User.Name,
		Email:       result.User.Email,
		AvatarUrl:   result.User.AvatarUrl,
		Residences:  make([]ResidenceResponse, len(*result.Residences)),
		Memberships: make([]MembershipResponse, len(*result.Memberships)),
	}

	for i, res := range *result.Residences {
		response.Residences[i] = ResidenceResponse{
			CondominiumName: res.CondominiumName,
			CondominiumID:   res.CondominiumID,
			Block:           res.Block,
			Number:          res.ApartementNumber,
			Type:            res.ResidentType,
			IsResponsible:   res.IsResponsible,
		}
	}

	for i, mem := range *result.Memberships {
		response.Memberships[i] = MembershipResponse{
			CondominiumName: mem.CondominiumName,
			CondominiumID:   mem.CondominiumID,
			Role:            mem.Role,
		}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, response)
}
