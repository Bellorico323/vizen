package controllers

import (
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
)

type RegisterDeviceHandler struct {
	RegisterDevice usecases.RegisterDeviceUseCase
}

type RegisterDeviceReq struct {
	FCMToken string `json:"fcmToken" validate:"required"`
	Platform string `json:"platform" validate:"oneof=android ios web"`
}

// @Summary Register Device token
// @Security BearerAuth
// @Tags Notifications
// @Accept 	json
// @Produce json
// @Param request body controllers.RegisterDeviceReq true "Register device payload"
// @Success 200 {object} common.SuccessResponse "Device registred successfully"
// @Failure 400 {object} common.ErrResponse "Invalid JSON or Request not pending"
// @Failure 422 {object} common.ValidationErrResponse "Validation failed"
// @Failure 401 {object} common.ErrResponse "User not authenticated"
// @Router /users/devices [post]
func (h *RegisterDeviceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[RegisterDeviceReq](r)
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

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	err = h.RegisterDevice.Exec(r.Context(), userID, data.FCMToken, data.Platform)
	if err != nil {
		slog.Error("Error while registring device", "error", err)

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "An unexpected error occurred while registring device",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, common.SuccessResponse{
		Message: "Device registred successfully",
	})
}
