package aggregates

import "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"

type UnloggedUser struct {
	*entities.UnloggedUser

	Plan   *entities.Plan
	Models []*entities.Model
}

func NewUnloggedUser(user *entities.UnloggedUser, plan *entities.Plan, models []*entities.Model) (*UnloggedUser, error) {
	err := plan.Validate()
	if err != nil {
		return nil, err
	}

	return &UnloggedUser{UnloggedUser: user, Plan: plan, Models: models}, nil
}
