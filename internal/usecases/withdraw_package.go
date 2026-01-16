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

type WithdrawPackageUC interface {
	Exec(ctx context.Context, req WithdrawPackageReq) error
}

type WithdrawPackageReq struct {
	PackageID     uuid.UUID
	UserID        uuid.UUID
	CondominiumID uuid.UUID
}

type WithdrawPackageUseCase struct {
	querier pgstore.Querier
}

func NewWithdrawPackageUseCase(q pgstore.Querier) *WithdrawPackageUseCase {
	return &WithdrawPackageUseCase{
		querier: q,
	}
}

var ErrPackageAlreadyWithdrawn = errors.New("package is already withdrawn")

func (uc *WithdrawPackageUseCase) Exec(ctx context.Context, req WithdrawPackageReq) error {
	packg, err := uc.querier.GetPackageById(ctx, req.PackageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPackageNotFound
		}

		return fmt.Errorf("Error while fetching package: %w", err)
	}

	if packg.CondominiumID != req.CondominiumID {
		return ErrPackageNotFound
	}

	if packg.Status == "withdrawn" {
		return ErrPackageAlreadyWithdrawn
	}

	_, err = uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})

	isStaff := false
	if err == nil {
		isStaff = true
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check staff permission: %w", err)
	}

	if !isStaff {
		userApts, err := uc.querier.GetApartmentsByUserId(ctx, pgstore.GetApartmentsByUserIdParams{
			CondominiumID: utils.ToPtr(req.CondominiumID),
			UserID:        req.UserID,
		})
		if err != nil {
			return fmt.Errorf("failed to check resident permission: %w", err)
		}

		isResidentOfTargetApt := false
		for _, apt := range userApts {
			if apt.ApartmentID == packg.ApartmentID {
				isResidentOfTargetApt = true
				break
			}
		}

		if !isResidentOfTargetApt {
			return ErrNoPermission
		}

	}

	err = uc.querier.UpdatePackageToWithdrawn(ctx, pgstore.UpdatePackageToWithdrawnParams{
		ID:          packg.ID,
		WithdrawnBy: utils.ToPtr(req.UserID),
	})
	if err != nil {
		return fmt.Errorf("Error while updating package: %w", err)
	}

	return nil
}
