package result

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"

type Refresh struct {
	AccessToken  model.Token
	RefreshToken model.Token
}
