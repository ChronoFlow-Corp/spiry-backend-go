package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/result"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/google/uuid"
)

type UseCaseRepository interface {
	SaveUser(ctx context.Context, user *aggregates.User) error
	GetUserByEmail(ctx context.Context, email string) (*aggregates.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*aggregates.User, error)
	GetBaseSubscription(ctx context.Context) (*entities.Subscription, error)
	GetModels(ctx context.Context) ([]*entities.Model, error)
	GetModelByName(ctx context.Context, name string) (*entities.Model, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error)
}

type OAuthProvider interface {
	GetAuthCodeURI(device string) string
	ExchangeCode(ctx context.Context, code string) (model.OauthInfo, error)
}

type TokenProvider interface {
	GeneratePair(userID, sessionID uuid.UUID) (model.Token, model.Token, error)
	ParseRefresh(token string) (model.ParsedToken, error)
}

type UseCase struct {
	repo          UseCaseRepository
	tokenProvider TokenProvider
	oAuthProvider OAuthProvider
}

func NewAuthUseCase(
	repo UseCaseRepository,
	tokenProvider TokenProvider,
	oAuthProvider OAuthProvider,
) *UseCase {
	return &UseCase{
		repo:          repo,
		tokenProvider: tokenProvider,
		oAuthProvider: oAuthProvider,
	}
}

func (uc *UseCase) Login(ctx context.Context, cm command.Login) (result.Login, error) {
	const op = "application.UseCase.Login"

	info, err := uc.oAuthProvider.ExchangeCode(ctx, cm.Code)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := uc.repo.GetUserByEmail(ctx, info.Email)
	if err != nil {
		var notFound *domain.ErrorNotFound
		if errors.As(err, &notFound) {
			return uc.newLogin(ctx, cm, info)
		}

		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	return uc.addLogin(ctx, cm, user)
}

func (uc *UseCase) GetAuthURI(device string) string {
	return uc.oAuthProvider.GetAuthCodeURI(device)
}

func (uc *UseCase) Refresh(ctx context.Context, cm command.Refresh) (result.Refresh, error) {
	const op = "application.UseCase.Refresh"

	parsed, err := uc.tokenProvider.ParseRefresh(cm.RefreshToken)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := uc.repo.GetUserByID(ctx, parsed.UserID)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	session, err := user.GetSessionByToken(cm.RefreshToken)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	access, refresh, err := uc.tokenProvider.GeneratePair(user.ID, session.ID)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	err = user.UpdateSessionToken(cm.RefreshToken, refresh.Token, refresh.ExpireAt)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	err = uc.repo.SaveUser(ctx, user)
	if err != nil {
		return result.Refresh{}, fmt.Errorf("%s: %w", op, err)
	}

	return result.Refresh{AccessToken: access, RefreshToken: refresh}, nil
}

func (uc *UseCase) LogOut(ctx context.Context, cm command.Logout) error {
	const op = "application.UseCase.Logout"

	user, err := uc.repo.GetUserByID(ctx, cm.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	user.RemoveSessionByID(cm.SessionID)

	err = uc.repo.SaveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (uc *UseCase) UserInfo(ctx context.Context) (result.UserInfo, error) {
	const op = "application.UseCase.UserInfo"

	userID, ok := models.GetUserIDFromCtx(ctx)
	if !ok {
		return result.UserInfo{}, fmt.Errorf("%s: %w", op, errors.New("user id not provided"))
	}

	u, err := uc.repo.GetUserByID(ctx, *userID)
	if err != nil {
		return result.UserInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	sub, err := uc.repo.GetSubscriptionByID(ctx, u.Plan.SubscriptionID)
	if err != nil {
		return result.UserInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return result.UserInfo{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Plan: result.Plan{
			ID:        u.Plan.ID,
			Name:      sub.Name,
			Price:     sub.Price,
			CreatedAt: u.Plan.CreatedAt,
			UpdatedAt: u.Plan.UpdatedAt,
		},
		Theme:     string(u.Theme),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (uc *UseCase) addLogin(
	ctx context.Context,
	cm command.Login,
	user *aggregates.User,
) (result.Login, error) {
	const op = "application.UseCase.addLogin"

	device, ok := cm.State["device"]
	if !ok {
		return result.Login{}, fmt.Errorf("%s: %w", op, errors.New("device not found"))
	}

	sessionID := uuid.New()

	access, refresh, err := uc.tokenProvider.GeneratePair(user.ID, sessionID)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	sessionEntity := entities.NewSession(
		time.Now(),
		refresh.ExpireAt,
		&user.ID,
		refresh.Token,
		device,
	)

	err = user.AddSession(sessionEntity)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	err = uc.repo.SaveUser(ctx, user)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	return result.Login{AccessToken: access, RefreshToken: refresh}, nil
}

func (uc *UseCase) newLogin(
	ctx context.Context,
	cm command.Login,
	info model.OauthInfo,
) (result.Login, error) {
	const op = "application.UseCase.newLogin"

	sub, err := uc.repo.GetBaseSubscription(ctx)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	userEntity := entities.NewUser(
		info.Email,
		info.Name,
		info.LastName,
		info.AvatarURL,
		entities.ThemeAuto,
		info.AccessToken,
		info.RefreshToken,
	)

	planEntity := entities.NewPlan(sub.Quote, sub.Level, sub.ID, userEntity.ID)

	sessionID := uuid.New()

	access, refresh, err := uc.tokenProvider.GeneratePair(userEntity.ID, sessionID)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	device, ok := cm.State["device"]
	if !ok {
		return result.Login{}, fmt.Errorf("%s: %w", op, errors.New("device not found"))
	}

	sessionEntity := entities.NewSession(
		time.Now(),
		refresh.ExpireAt,
		&userEntity.ID,
		refresh.Token,
		device,
	)
	sessionEntity.ID = sessionID

	user, err := aggregates.NewUser(
		userEntity,
		planEntity,
		nil,
		[]*entities.Session{sessionEntity},
	)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	err = uc.repo.SaveUser(ctx, user)
	if err != nil {
		return result.Login{}, fmt.Errorf("%s: %w", op, err)
	}

	return result.Login{AccessToken: access, RefreshToken: refresh}, nil
}
