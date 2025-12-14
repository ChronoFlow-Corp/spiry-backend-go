package llm

import (
	"cmp"
	"slices"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

func lightestModel(models []*entities.Model) *entities.Model {
	return slices.MinFunc(models, func(a, b *entities.Model) int {
		return cmp.Compare(a.MinLevel, b.MinLevel)
	})
}
