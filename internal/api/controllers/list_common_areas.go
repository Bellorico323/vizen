package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListCommonAreasHandler struct {
	ListCommonAreas usecases.ListCommonAreasUC
}

type ListCommonAreasResponse struct {
	Data []pgstore.CommonArea `json:"data"`
}

// List handles listing of common areas
// @Summary      List Common Areas
// @Description  Lists all common areas available in the condominium. Open to all members.
// @Tags         Common Areas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        condominiumId query     string  true   "Condominium UUID"
// @Success      200     {object}  controllers.ListCommonAreasResponse
// @Failure      403     {object}  common.ErrResponse
// @Router       /common_areas [get]
func (h *ListCommonAreasHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "Unauthorized",
		})
		return
	}

	condoIdStr := r.URL.Query().Get("condominiumId")
	condoID, err := uuid.Parse(condoIdStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid condominiumId",
		})
		return
	}

	areas, err := h.ListCommonAreas.Exec(r.Context(), usecases.ListCommonAreasReq{
		UserID:        userID,
		CondominiumID: condoID,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{Message: "You do not have access to this condominium"})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{Message: "Failed to list common areas"})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, ListCommonAreasResponse{
		Data: areas,
	})
}
