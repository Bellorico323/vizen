package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PackageListItem struct {
	ID              uuid.UUID `json:"id"`
	RecipientName   *string   `json:"recipientName"`
	Status          string    `json:"status"`
	ReceivedAt      time.Time `json:"receivedAt"`
	Block           *string   `json:"block,omitempty"`
	ApartmentNumber *string   `json:"apartmentNumber,omitempty"`
	ReceivedByName  *string   `json:"receivedByName,omitempty"`
	PhotoUrl        *string   `json:"photoUrl,omitempty"`
}

type ListPackagesUC interface {
	Exec(ctx context.Context, req ListPackagesReq) ([]PackageListItem, error)
}

type ListPackagesReq struct {
	UserID        uuid.UUID
	CondominiumID uuid.UUID
	ApartmentID   *uuid.UUID
	Status        *string
}

type ListPackagesUseCase struct {
	querier pgstore.Querier
}

func NewListPackagesUseCase(q pgstore.Querier) *ListPackagesUseCase {
	return &ListPackagesUseCase{querier: q}
}

func (uc *ListPackagesUseCase) Exec(ctx context.Context, req ListPackagesReq) ([]PackageListItem, error) {
	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})

	var isAdmin bool

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			isAdmin = false
		} else {
			return nil, err
		}
	}

	isAdmin = role == "admin" || role == "syndic"

	if req.ApartmentID != nil {
		if !isAdmin {
			userApts, err := uc.querier.GetApartmentsByUserId(ctx, pgstore.GetApartmentsByUserIdParams{
				UserID:        req.UserID,
				CondominiumID: utils.ToPtr(req.CondominiumID),
			})
			if err != nil {
				return nil, err
			}

			isResident := false
			for _, apt := range userApts {
				if apt.ApartmentID == *req.ApartmentID {
					isResident = true
					break
				}
			}
			if !isResident {
				return nil, ErrNoPermission
			}
		}

		rows, err := uc.querier.ListPackagesByApartment(ctx, pgstore.ListPackagesByApartmentParams{
			ApartmentID: *req.ApartmentID,
			Status:      req.Status,
		})
		if err != nil {
			return nil, err
		}

		items := make([]PackageListItem, len(rows))
		for i, r := range rows {
			items[i] = PackageListItem{
				ID:             r.ID,
				RecipientName:  r.RecipientName,
				Status:         r.Status,
				ReceivedAt:     r.ReceivedAt,
				PhotoUrl:       r.PhotoUrl,
				ReceivedByName: &r.ReceivedByName,
			}
		}
		return items, nil
	}

	if !isAdmin {
		return nil, ErrNoPermission
	}

	rows, err := uc.querier.ListPackagesByCondominium(ctx, pgstore.ListPackagesByCondominiumParams{
		CondominiumID: req.CondominiumID,
		Status:        req.Status,
	})
	if err != nil {
		return nil, err
	}

	items := make([]PackageListItem, len(rows))
	for i, r := range rows {
		items[i] = PackageListItem{
			ID:              r.ID,
			RecipientName:   r.RecipientName,
			Status:          r.Status,
			ReceivedAt:      r.ReceivedAt,
			Block:           r.Block,
			ApartmentNumber: &r.ApartmentNumber,
		}
	}

	return items, nil
}
