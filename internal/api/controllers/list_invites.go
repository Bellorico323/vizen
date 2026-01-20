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

type ListInvitesHandler struct {
	ListInvites usecases.ListInvitesUC
}

type ListInvitesResponse struct {
	Data []struct {
		ID              uuid.UUID `json:"id"`
		GuestName       string    `json:"guestName"`
		GuestType       string    `json:"guestType"`
		StartsAt        string    `json:"startsAt"`
		EndsAt          string    `json:"endsAt"`
		RevokedAt       *string   `json:"revokedAt,omitempty"`
		Token           uuid.UUID `json:"token"`
		ApartmentBlock  string    `json:"block"`
		ApartmentNumber string    `json:"apartmentNumber"`
	} `json:"data"`
}

// Handle lists invites
// @Summary      List Invites
// @Description  Get a list of invites. Residents see only their apartment's invites. Staff can see all.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        condominiumId query     string  true   "Condominium UUID"
// @Param        apartmentId   query     string  false  "Apartment UUID (Required for residents)"
// @Param        onlyActive    query     boolean false  "If true, returns only future/active invites"
// @Param        page          query     int     false  "Page number (default 1)"
// @Param        limit         query     int     false  "Items per page (default 20)"
// @Success      200     {object}  ListInvitesResponse
// @Failure      403     {object}  common.ErrResponse
// @Router       /invites [get]
func (h *ListInvitesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{Message: "Unauthorized"})
		return
	}

	q := r.URL.Query()
	condoID, err := uuid.Parse(q.Get("condominiumId"))
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid condominiumId"})
		return
	}
	var aptID *uuid.UUID
	if idStr := q.Get("apartmentId"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			aptID = &id
		}
	}

	onlyActive := q.Get("onlyActive") == "true"

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	rows, err := h.ListInvites.Exec(r.Context(), usecases.ListInvitesReq{
		CondominiumID: condoID,
		UserID:        userID,
		ApartmentID:   aptID,
		OnlyActive:    onlyActive,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})

	if err != nil {
		if errors.Is(err, usecases.ErrNoPermission) {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{Message: "Permission denied"})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{Message: "Failed to list invites"})
		return
	}

	resp := ListInvitesResponse{}
	resp.Data = make([]struct {
		ID              uuid.UUID `json:"id"`
		GuestName       string    `json:"guestName"`
		GuestType       string    `json:"guestType"`
		StartsAt        string    `json:"startsAt"`
		EndsAt          string    `json:"endsAt"`
		RevokedAt       *string   `json:"revokedAt,omitempty"`
		Token           uuid.UUID `json:"token"`
		ApartmentBlock  string    `json:"block"`
		ApartmentNumber string    `json:"apartmentNumber"`
	}, len(rows))

	for i, row := range rows {
		var revokedAt *string
		if row.RevokedAt != nil {
			s := row.RevokedAt.Format(time.RFC3339)
			revokedAt = &s
		}

		resp.Data[i].ID = row.ID
		resp.Data[i].GuestName = row.GuestName
		resp.Data[i].GuestType = row.GuestType
		resp.Data[i].StartsAt = row.StartsAt.Format(time.RFC3339)
		resp.Data[i].EndsAt = row.EndsAt.Format(time.RFC3339)
		resp.Data[i].RevokedAt = revokedAt
		resp.Data[i].Token = row.Token
		if row.Block != nil {
			resp.Data[i].ApartmentBlock = *row.Block
		}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, resp)
}
