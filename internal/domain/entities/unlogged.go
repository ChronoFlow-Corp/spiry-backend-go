package entities

import "github.com/google/uuid"

type UnloggedUser struct {
	ID uuid.UUID
	IP string

	PlanID uuid.UUID
}

func NewUnloggedUser(ip string, quote ModalitiesQuote) *UnloggedUser {
	return &UnloggedUser{
		ID: uuid.New(),
		IP: ip,
	}
}
