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
	"github.com/jackc/pgx/v5/pgconn"
)

type ApproveAccessRequestHandler struct {
	ApproveAccessRequest usecases.ApproveAccessRequestUC
}

type ApproveAccessRequestReq struct {
	AccessRequestID uuid.UUID `json:"accessRequestId" validate:"required"`
	IsResponsible   bool      `json:"isResponsible"`
}

type ApproveAccessRequestResponse struct {
	Message string `json:"message"`
}

// Handle approves a pending access request
// @Summary 		Approve Access Request
// @Description	Approves a pending request, turning the user into an official resident.
// @Security		BearerAuth
// @Tags			Access Requests
// @Accept			json
// @Produce			json
// @Param			request body controllers.ApproveAccessRequestReq true "Approval payload"
// @Success			200 {object} controllers.ApproveAccessRequestResponse "Request approved successfully"
// @Failure 		400	{object} common.ErrResponse "Invalid JSON or Request not pending"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			403 {object} common.ErrResponse "User does not have permission"
// @Failure			404 {object} common.ErrResponse "Access request not found"
// @Failure			409 {object} common.ErrResponse "User is already a resident"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/access-requests/approve [post]
func (h *ApproveAccessRequestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[ApproveAccessRequestReq](r)
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

	payload := usecases.ApproveAccessRequestReq{
		ReviewerID:      reviewerId,
		AccessRequestID: data.AccessRequestID,
		IsResponsible:   data.IsResponsible,
	}

	err = h.ApproveAccessRequest.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while approving access request", "error", err)

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
				Message: "You do not have permission to approve requests for this condominium",
			})

		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
					Message: "User is already a resident of this apartment",
				})
				return
			}

			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while approving request",
			})
		}
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, ApproveAccessRequestResponse{
		Message: "Access request approved successfully",
	})
}
