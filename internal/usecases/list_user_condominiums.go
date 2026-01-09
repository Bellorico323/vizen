package usecases

import (
	"context"
	"sort"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type ListUserCondominiumsUC interface {
	Exec(ctx context.Context, userID uuid.UUID) ([]UserCondominiumDTO, error)
}

type UserCondominiumDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Cnpj    string    `json:"cnpj"`
	Address string    `json:"address"`
	Role    string    `json:"role"`
}

type ListUserCondominiumsUseCase struct {
	querier pgstore.Querier
}

func NewListUserCondominiumsUseCase(q pgstore.Querier) *ListUserCondominiumsUseCase {
	return &ListUserCondominiumsUseCase{querier: q}
}

func (uc *ListUserCondominiumsUseCase) Exec(ctx context.Context, userID uuid.UUID) ([]UserCondominiumDTO, error) {
	rows, err := uc.querier.ListCondominiunsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	uniqueCondos := make(map[uuid.UUID]UserCondominiumDTO)

	for _, row := range rows {
		existing, found := uniqueCondos[row.ID]

		if !found {
			uniqueCondos[row.ID] = UserCondominiumDTO{
				ID:      row.ID,
				Name:    row.Name,
				Cnpj:    row.Cnpj,
				Address: row.Address,
				Role:    row.RoleName,
			}
			continue
		}

		if isHigherRole(row.RoleName, existing.Role) {
			existing.Role = row.RoleName
			uniqueCondos[row.ID] = existing
		}
	}

	result := make([]UserCondominiumDTO, 0, len(uniqueCondos))
	for _, condo := range uniqueCondos {
		result = append(result, condo)
	}

	sort.Slice(result, func(i, j int) bool {
		p1 := getRolePriority(result[i].Role)
		p2 := getRolePriority(result[j].Role)

		if p1 == p2 {
			return result[i].Name < result[j].Name
		}

		return p1 > p2
	})

	return result, nil
}

func getRolePriority(role string) int {
	switch role {
	case "syndic":
		return 3
	case "admin":
		return 2
	default: // resident
		return 1
	}
}

func isHigherRole(newRole, currentRole string) bool {
	return getRolePriority(newRole) > getRolePriority(currentRole)
}
