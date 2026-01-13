package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListUserApartmentsHandler struct {
	ListUserApartments usecases.ListUserApartmentsUC
}

type UserApartmentResponse struct {
	ApartmentID     uuid.UUID `json:"apartmentId"`
	Block           *string   `json:"block"`
	Number          string    `json:"number"`
	Type            string    `json:"type"`
	IsResponsible   bool      `json:"isResponsible"`
	CondominiumName string    `json:"condominiumName"`
	CondominiumID   uuid.UUID `json:"condominiumId"`
}

// Handle lists user apartments
// @Summary 		List User Apartments
// @Description	Get all apartments belonging to the authenticated user. Optionally filter by condominium.
// @Security		BearerAuth
// @Tags			Users
// @Produce			json
// @Param			condominiumId query string false "Condominium UUID (Optional)"
// @Success			200 {object} []controllers.UserApartmentResponse
// @Failure			400 {object} common.ErrResponse "Invalid UUID format"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/users/apartments [get]
func (h *ListUserApartmentsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	query := r.URL.Query()
	condoParam := query.Get("condominiumId")

	var condoID *uuid.UUID

	if condoParam != "" {
		parsed, err := uuid.Parse(condoParam)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: "Invalid condominiumId format",
			})
			return
		}
		condoID = &parsed
	}

	data, err := h.ListUserApartments.Exec(r.Context(), usecases.ListUserApartmentsReq{
		UserID:        userID,
		CondominiumID: condoID,
	})

	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "User not found",
			})
			return
		}

		if errors.Is(err, usecases.ErrUserNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "Failed to fetch apartments",
			})
			return
		}
	}

	response := make([]UserApartmentResponse, len(data))

	for i, item := range data {
		response[i] = UserApartmentResponse{
			ApartmentID:     item.ApartmentID,
			Block:           item.Block,
			Number:          item.Number,
			Type:            item.Type,
			IsResponsible:   item.IsResponsible,
			CondominiumName: item.CondominiumName,
			CondominiumID:   item.CondominiumID,
		}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, response)
}
