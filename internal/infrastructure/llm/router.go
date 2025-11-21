package llm

import (
	"cmp"
	"slices"
	"sync"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type router struct {
	models map[uuid.UUID]modelWithMetrics
}

type modelWithMetrics struct {
	model        entities.Model
	tks          float64
	responseTime time.Duration
	mu           *sync.RWMutex
}

func lightestModel(models []*entities.Model) *entities.Model {
	return slices.MinFunc(models, func(a, b *entities.Model) int {
		return cmp.Compare(a.MinLevel, b.MinLevel)
	})
}

//func modalityMatch(
//	models []*entities.Model,
//	modalities []entities.Modality,
//) (*entities.Model, error) {
//	match := make([]*entities.Model, 0, len(models))
//	for _, model := range models {
//		if slices.Equal(model.Modalities, modalities) {
//			match = append(match, model)
//		}
//	}
//
//	if len(match) == 0 {
//		return nil, errors.New("models do not match")
//	}
//
//	return lightestModel(match), nil
//}
