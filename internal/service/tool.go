package service

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
)

type toolProvider interface{
	Get(ctx context.Context) ([]entities.Tool, error)
}
type Tool struct{

}
