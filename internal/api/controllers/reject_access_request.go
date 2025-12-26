package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type RejectAccessRequestHandler struct {
	RejectAccessRequest usecases.RejectAccessRequestUC
}

type RejectAccessRequestReq struct {
	AccessRequestID uuid.UUID `json:"accessRequestId" validate:"required"`
}

type RejectAccessRequestResponse struct {
	Message string `json:"message"`
}

// Handle rejects a pending access request
// @Summary 		Reject Access Request
// @Description	Rejects a pending request, denying the user access to the condominium.
// @Security		BearerAuth
// @Tags			Access Requests
// @Accept			json
// @Produce			json
// @Param			request body controllers.RejectAccessRequestReq true "Rejection payload"
// @Success			200 {object} controllers.RejectAccessRequestResponse "Request rejected successfully"
// @Failure 		400	{object} common.ErrResponse "Invalid JSON or Request not pending"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			403 {object} common.ErrResponse "User does not have permission"
// @Failure			404 {object} common.ErrResponse "Access request not found"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/access-requests/reject [post]
func (h *RejectAccessRequestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[RejectAccessRequestReq](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid JSON payload",
		})
		return
	}

	if validationErrors := validator.ValidateStruct(data); len(validationErrors) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, common.ValidationErrResponse{
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}

	reviewerId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	payload := usecases.RejectAccessRequestReq{
		ReviewerID:      reviewerId,
		AccessRequestID: data.AccessRequestID,
	}

	err = h.RejectAccessRequest.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while rejecting access request", "error", err)

		switch {
		case errors.Is(err, usecases.ErrAccessRequestNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Access request not found",
			})

		case errors.Is(err, usecases.ErrRequestNotPending):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: "Access request is not pending",
			})

		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to reject requests for this condominium",
			})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while rejecting request",
			})
		}
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, RejectAccessRequestResponse{
		Message: "Access request rejected successfully",
	})
}
