package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetPackageUC interface {
	Exec(ctx context.Context, req GetPackageReq) (pgstore.GetPackageByIdRow, error)
}

type GetPackageReq struct {
	UserID    uuid.UUID
	PackageID uuid.UUID
}

type GetPackageUseCase struct {
	querier pgstore.Querier
}

func NewGetPackageUseCase(q pgstore.Querier) *GetPackageUseCase {
	return &GetPackageUseCase{
		querier: q,
	}
}

var ErrPackageNotFound = errors.New("package not found")

func (uc *GetPackageUseCase) Exec(ctx context.Context, req GetPackageReq) (pgstore.GetPackageByIdRow, error) {
	pkg, err := uc.querier.GetPackageById(ctx, req.PackageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.GetPackageByIdRow{}, ErrPackageNotFound
		}
		return pgstore.GetPackageByIdRow{}, fmt.Errorf("failed to get package: %w", err)
	}

	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: pkg.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.GetPackageByIdRow{}, ErrNoPermission
		}
		return pgstore.GetPackageByIdRow{}, err
	}

	if role == "admin" || role == "syndic" || role == "doorman" {
		return pkg, nil
	}

	userApartments, err := uc.querier.GetApartmentsByUserId(ctx, pgstore.GetApartmentsByUserIdParams{
		UserID:        req.UserID,
		CondominiumID: utils.ToPtr(pkg.CondominiumID),
	})
	if err != nil {
		return pgstore.GetPackageByIdRow{}, err
	}

	isResidentOfTargetApartment := false
	for _, apt := range userApartments {
		if apt.ApartmentID == pkg.ApartmentID {
			isResidentOfTargetApartment = true
			break
		}
	}

	if !isResidentOfTargetApartment {
		return pgstore.GetPackageByIdRow{}, ErrNoPermission
	}

	return pkg, nil
}
