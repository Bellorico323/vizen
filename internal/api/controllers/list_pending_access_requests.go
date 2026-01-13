package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListPendingAccessRequestHandler struct {
	ListPendingAccessRequests usecases.ListPendingAccessRequestUC
}

type AccessRequestResponse struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UserName  string    `json:"userName"`
	UserEmail string    `json:"userEmail"`
}

// Handle lists pending access requests
// @Summary 		List Pending Access Requests
// @Description	Returns all pending requests for a condominium. Only for Admins/Syndics.
// @Security		BearerAuth
// @Tags			Access Requests
// @Produce			json
// @Param			condominiumId query string true "Condominium UUID"
// @Success			200 {object} []controllers.AccessRequestResponse
// @Failure			400 {object} common.ErrResponse "Invalid UUID"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			403 {object} common.ErrResponse "Permission denied"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/access-requests/pending [get]
func (h *ListPendingAccessRequestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	condoIDStr := r.URL.Query().Get("condominiumId")
	if condoIDStr == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "condominiumId is required",
		})
		return
	}

	condoID, err := uuid.Parse(condoIDStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid condominiumId format",
		})
		return
	}

	data, err := h.ListPendingAccessRequests.Exec(r.Context(), usecases.ListPendingAccessRequestReq{
		UserID:        userID,
		CondominiumID: condoID,
	})

	if err != nil {
		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to view requests for this condominium",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to fetch requests",
		})
		return
	}

	response := make([]AccessRequestResponse, len(data))
	for i, item := range data {
		response[i] = AccessRequestResponse{
			ID:        item.ID,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			UserName:  item.Name,
			UserEmail: item.Email,
		}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, response)
}
