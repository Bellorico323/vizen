package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WithdrawPackageHandler struct {
	WithdrawPackage usecases.WithdrawPackageUC
}

type WithdrawPackageRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId"`
}

// Handle marks a package as withdrawn
// @Summary      Withdraw Package
// @Description  Marks a package as delivered/withdrawn. Can be performed by Condo Staff (for any package) or the Package Owner (Resident).
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                           true  "Package UUID"
// @Param        request body      controllers.WithdrawPackageRequest true  "Condominium ID"
// @Success      204     "No Content"
// @Failure      400     {object}  common.ErrResponse             "Invalid ID or Payload"
// @Failure      401     {object}  common.ErrResponse             "User not authenticated"
// @Failure      403     {object}  common.ErrResponse             "Permission denied"
// @Failure      404     {object}  common.ErrResponse             "Package not found"
// @Failure      409     {object}  common.ErrResponse             "Package already withdrawn"
// @Failure      500     {object}  common.ErrResponse             "Internal server error"
// @Router       /packages/{id}/withdraw [patch]
func (h *WithdrawPackageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	packageID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid package ID format",
		})
		return
	}

	data, err := jsonutils.DecodeJson[WithdrawPackageRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid JSON payload",
		})
		return
	}

	err = h.WithdrawPackage.Exec(r.Context(), usecases.WithdrawPackageReq{
		PackageID:     packageID,
		UserID:        userID,
		CondominiumID: data.CondominiumID,
	})

	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrPackageNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Package not found",
			})
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to withdraw this package",
			})
		case errors.Is(err, usecases.ErrPackageAlreadyWithdrawn):
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
				Message: "Package has already been withdrawn",
			})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "Failed to withdraw package",
			})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
