package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type ValidateInviteHandler struct {
	ValidateInvite usecases.ValidateInviteUC
}

type ValidateInviteRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	Token         uuid.UUID `json:"token" validate:"required"`
}

type ValidateInviteResponse struct {
	Message string                      `json:"message"`
	Invite  pgstore.GetInviteByTokenRow `json:"invite"`
}

// Handle validates an access token
// @Summary      Validate Access Token (QR Code)
// @Description  Validates a visitor's QR code token. Checks if the token exists, belongs to the condominium, is within the valid time window, and has not been revoked. Records the entry in access logs upon success.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      controllers.ValidateInviteRequest true "Token Validation Data"
// @Success      200     {object}  controllers.ValidateInviteResponse
// @Failure      400     {object}  common.ErrResponse            "Invalid Payload"
// @Failure      401     {object}  common.ErrResponse            "User not authenticated"
// @Failure      403     {object}  common.ErrResponse            "Business Rule Violation (Expired, Revoked, Not Started) or Permission Denied"
// @Failure      404     {object}  common.ErrResponse            "Invite/Token not found"
// @Failure      422     {object}  common.ValidationErrResponse  "Validation Failed"
// @Failure      500     {object}  common.ErrResponse            "Internal server error"
// @Router       /invites/validate [post]
func (h *ValidateInviteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[ValidateInviteRequest](r)
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

	userId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	payload := usecases.ValidateInviteReq{
		CondominiumID: data.CondominiumID,
		Token:         data.Token,
		UserID:        userId,
	}

	invite, err := h.ValidateInvite.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while validating invite",
			"error", err,
			"condominiumId", data.CondominiumID,
			"token", data.Token,
			"userId", userId,
		)

		switch {
		case errors.Is(err, usecases.ErrInviteNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to validate invites in this condominium",
			})

		case errors.Is(err, usecases.ErrInviteAlreadyRevoked):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrInviteNotStarted):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrInviteExpired):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while validating invite",
			})
		}

		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, ValidateInviteResponse{
		Message: "Access granted",
		Invite:  invite,
	})
}
