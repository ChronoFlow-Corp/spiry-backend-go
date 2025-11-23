package main

import (
	"log/slog"
	"net/http"

	_ "github.com/ChronoFlow-Corp/spiry-backend-go/docs"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/auth"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/auth/connect"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/auth/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/llm"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/pgx"
	authRepository "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/repository/auth"
	chattingRepository "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/repository/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/chats"
	chatsStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/chats/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands"
	commandsStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands/pgx"
	commandsmedia "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias"
	commandsMediaStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models"
	modelsStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans"
	plansStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans/pgx"
	resultmedias "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/result_medias"
	resultsMediasStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/result_medias/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results"
	resultsStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/sessions"
	sessionStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/sessions/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions"
	subStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/subscriptions/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/tools"
	toolsStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/tools/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/users"
	userStorage "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/users/pgx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/handlers"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(rest.RouteModule)),
		fx.ResultTags(`group:"routes"`),
	)
}

// main godoc
//
//	@title		Spirx API
//	@version	1.0
//	@BasePath	/api
func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			pgx.NewPool,
			pgx.NewManager,
			fx.Annotate(
				plansStorage.NewPgx,
				fx.As(new(plans.PlanStorage)),
			),
			fx.Annotate(
				subStorage.NewPgx,
				fx.As(new(subscriptions.SubscriptionStorage)),
			),

			fx.Annotate(
				userStorage.NewPgx,
				fx.As(new(users.UserStorage)),
			),
			fx.Annotate(
				sessionStorage.NewPgx,
				fx.As(new(sessions.SessionStorage)),
			),

			fx.Annotate(
				modelsStorage.NewPgx,
				fx.As(new(models.ModelStorage)),
			),
			fx.Annotate(
				chatsStorage.NewPgx,
				fx.As(new(chats.ChatStorage)),
			),
			fx.Annotate(
				toolsStorage.NewPgx,
				fx.As(new(tools.ToolStorage)),
			),
			fx.Annotate(
				commandsStorage.NewPgx,
				fx.As(new(commands.CommandStorage)),
			),
			fx.Annotate(
				commandsMediaStorage.NewPgx,
				fx.As(new(commandsmedia.CommandMediaStorage)),
			),
			fx.Annotate(
				resultsMediasStorage.NewPgx,
				fx.As(new(resultmedias.ResultMediaStorage)),
			),

			fx.Annotate(
				resultsStorage.NewPgx,
				fx.As(new(results.ResultStorage)),
			),
			fx.Annotate(
				authRepository.NewAuthRepository,
				fx.As(new(auth.UseCaseRepository)),
				fx.As(new(chatting.AuthRepository)),
			),

			fx.Annotate(
				connect.NewGoogleOauthFx,
				fx.As(new(auth.OAuthProvider)),
			),

			fx.Annotate(
				jwt.NewFx,
				fx.As(new(auth.TokenProvider)),
				fx.As(new(handlers.JWTProvider)),
			),

			fx.Annotate(
				auth.NewAuthUseCase,
				fx.As(new(auth.Service)),
			),

			fx.Annotate(
				chattingRepository.NewRepository,
				fx.As(new(chatting.UseCaseRepository)),
			),

			fx.Annotate(llm.NewOpenRouter, fx.As(new(service.LLmClient))),
			fx.Annotate(service.NewChatting),
			fx.Annotate(chatting.NewChatting, fx.As(new(chatting.Service))),

			AsRoute(handlers.NewAuthModule),
			AsRoute(handlers.NewChattingModule),

			fx.Annotate(
				rest.NewServeMux,
				fx.As(new(chi.Router)),
				fx.ParamTags(`group:"routes"`),
			),

			slog.Default,

			fx.Annotate(
				rest.NewHTTPServer,
			),
		),
		fx.Invoke(
			func(srv *http.Server) {}),
	)

	app.Run()
}
