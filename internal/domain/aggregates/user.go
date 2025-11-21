package aggregates

import (
	"slices"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

const (
	maxSessions = 5
)

type User struct {
	*entities.User

	Plan     *entities.Plan
	Sessions []*entities.Session

	// Models is allowed model according to plan.
	// Models will be calculated each time they are received from the repository.
	Models []*entities.Model
}

func NewUser(
	user *entities.User,
	plan *entities.Plan,
	models []*entities.Model,
	sessions []*entities.Session) (*User, error) {
	const op = "domain.aggregates.NewUser"

	err := user.Validate()
	if err != nil {
		return nil, err
	}

	err = plan.Validate()
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		err = session.Validate()
		if err != nil {
			return nil, err
		}
	}

	return &User{User: user, Plan: plan, Sessions: sessions, Models: models}, nil
}

func (u *User) AddSession(session *entities.Session) error {
	err := session.Validate()
	if err != nil {
		return err
	}

	if u.Sessions == nil {
		u.Sessions = make([]*entities.Session, 0)
	}

	if len(u.Sessions) >= maxSessions {
		return domain.NewValidationError(nil, "sessions", "too many sessions")
	}

	u.Sessions = append(u.Sessions, session)

	return nil
}

func (u *User) UpdateSessionToken(old string, new string, expiresAt time.Time) error {
	for i, session := range u.Sessions {
		if session.Token == old {
			u.Sessions[i].Token = new
			u.Sessions[i].UpdatedAt = time.Now().UTC()
			u.Sessions[i].ExpiresAt = expiresAt
			u.Sessions[i].LastLogin = time.Now()

			err := session.Validate()
			if err != nil {
				return err
			}

			return nil
		}
	}

	return domain.NewNotFound(nil, "not found session", "session", old)
}

func (u *User) GetSessionByToken(token string) (*entities.Session, error) {
	for _, session := range u.Sessions {
		if session.Token == token {
			return session, nil
		}
	}

	return nil, domain.NewNotFound(nil, "not found session", "session", token)
}

func (u *User) RemoveSessionByID(id uuid.UUID) {
	for i, session := range u.Sessions {
		if session.ID == id {
			u.Sessions = slices.Delete(u.Sessions, i, i)
		}
	}
}
