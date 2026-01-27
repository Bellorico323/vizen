package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateBillUC interface {
	Exec(ctx context.Context, req CreateBillReq) (pgstore.Bill, error)
}

type CreateBillReq struct {
	UserID        uuid.UUID
	CondominiumID uuid.UUID
	ApartmentID   uuid.UUID
	BillType      string
	ValueInCents  int64
	DueDate       time.Time
	DigitableLine *string
	PixCode       *string
}

type CreateBillUseCase struct {
	querier pgstore.Querier
}

func NewCreateBillUseCase(q pgstore.Querier) *CreateBillUseCase {
	return &CreateBillUseCase{
		querier: q,
	}
}

var (
	ErrInvalidBillType    = errors.New("Invalid bill type")
	ErrInvalidBillDueDate = errors.New("Invalid due date")
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
)

func (uc *CreateBillUseCase) Exec(ctx context.Context, req CreateBillReq) (pgstore.Bill, error) {
	if validBillType := isValidBillType(req.BillType); !validBillType {
		return pgstore.Bill{}, ErrInvalidBillType
	}

	if req.ValueInCents <= 0 {
		return pgstore.Bill{}, ErrInvalidAmount
	}

	dueDateNormalized := time.Date(req.DueDate.Year(), req.DueDate.Month(), req.DueDate.Day(), 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if dueDateNormalized.Before(today) {
		return pgstore.Bill{}, ErrInvalidBillDueDate
	}

	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bill{}, ErrNoPermission
		}
		return pgstore.Bill{}, fmt.Errorf("Error while validating user role")
	}

	if role != "admin" && role != "syndic" {
		return pgstore.Bill{}, ErrNoPermission
	}

	if req.DueDate.Before(time.Now()) {
		return pgstore.Bill{}, ErrInvalidBillDueDate
	}

	apartment, err := uc.querier.GetApartmentById(ctx, req.ApartmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bill{}, ErrApartmentNotFound
		}
		return pgstore.Bill{}, fmt.Errorf("Error while fetching apartment: %w", err)
	}

	if apartment.CondominiumID != req.CondominiumID {
		return pgstore.Bill{}, ErrApartmentIsNotFromCondominium
	}

	var cleanDigitableLine *string
	if req.DigitableLine != nil {
		clean := sanitizeNumbers(*req.DigitableLine)
		if clean != "" {
			cleanDigitableLine = &clean
		}
	}

	bill, err := uc.querier.CreateBill(ctx, pgstore.CreateBillParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   req.ApartmentID,
		BillType:      req.BillType,
		ValueInCents:  req.ValueInCents,
		DueDate:       dueDateNormalized,
		DigitableLine: cleanDigitableLine,
		PixCode:       req.PixCode,
	})
	if err != nil {
		return pgstore.Bill{}, fmt.Errorf("Error while creating bill: %w", err)
	}

	return bill, nil
}

func isValidBillType(billType string) bool {
	switch billType {
	case "rent", "condominium_fee", "water", "electricity", "gas", "fine":
		return true

	default:
		return false
	}
}

func sanitizeNumbers(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}
