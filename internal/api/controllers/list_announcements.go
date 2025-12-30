package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListAnnouncementsHandler struct {
	ListAnnouncements usecases.ListAnnouncementsUC
}

type AnnouncementResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	AuthorName *string   `json:"authorName"`
}

// Handle lists announcements
// @Summary 		List Announcements
// @Description	Get a paginated list of announcements for a specific condominium.
// @Security		BearerAuth
// @Tags			Announcements
// @Produce			json
// @Param			condominiumId query string true "Condominium UUID"
// @Param			page query int false "Page number (default 1)"
// @Param			limit query int false "Items per page (default 10)"
// @Success			200 {object} []controllers.AnnouncementResponse
// @Failure 		400	{object} common.ErrResponse "Invalid UUID or parameters"
// @Failure			403 {object} common.ErrResponse "User does not have permission"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/announcements [get]
func (h *ListAnnouncementsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	query := r.URL.Query()

	condoIDStr := query.Get("condominiumId")
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

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	req := usecases.ListAnnouncementsReq{
		CondominiumID: condoID,
		UserID:        userId,
		Page:          int32(page),
		Limit:         int32(limit),
	}

	data, err := h.ListAnnouncements.Exec(r.Context(), req)
	if err != nil {
		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You are not a member of this condominium.",
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to fetch announcements",
		})
		return
	}

	response := make([]AnnouncementResponse, len(data))
	for i, item := range data {
		response[i] = AnnouncementResponse{
			ID:         item.ID,
			Title:      item.Title,
			Content:    item.Content,
			CreatedAt:  item.CreatedAt,
			AuthorName: item.AuthorName,
		}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, response)
}
