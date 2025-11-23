package result

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"

type Login struct {
	AccessToken  model.Token
	RefreshToken model.Token
}
