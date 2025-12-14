package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/auth"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/middlewares"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/pkg"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthModule struct {
	service     auth.Service
	j           JWTProvider
	frontendURL *url.URL
}

func NewAuthModule(cfg *config.Config, service auth.Service, j JWTProvider) *AuthModule {
	u, err := url.Parse(fmt.Sprintf("http://"))
	if err != nil {
		panic("invalid frontend URL " + err.Error())
	}

	return &AuthModule{
		service:     service,
		j:           j,
		frontendURL: u,
	}
}

func (a *AuthModule) Register(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Route("/connect", func(r chi.Router) {
			r.Get("/google", a.Login)
			r.Get("/google/callback", a.Callback)
		})

		r.Get("/refresh", a.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthJwt(a.j))
			r.Get("/user-info", a.UserInfo)
			r.Get("/logout", a.Logout)
		})
	})
}

// Login godoc
//
//	@Summary	redirect to google oAuth URL
//	@Tags		auth
//	@Param		User-Agent	header	string	true	"<device-name>"	example("PostmanRuntime/7.48.0")
//	@Success	307
//	@Header		307	{string}	Set-Cookie	"Set refresh token cookie; e.g. refresh_token=<token>; HttpOnly; Path=/api/auth/refresh; Secure"
//	@Header		307	{string}	Set-Cookie	"Set access token cookie; e.g. access_token=<token>; HttpOnly; Path=/; Secure"
//	@Router		/api/auth/connect/google [get]
func (a *AuthModule) Login(w http.ResponseWriter, r *http.Request) {
	agent := r.Header.Get("User-Agent")
	if agent == "" {
		slctx.Logger(r.Context()).Debug("no user agent")
		agent = r.Host
	}
	uri := a.service.GetAuthURI(agent)

	http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
}

// Callback godoc
//
//	@Summary	google redirect to this route
//	@Tags		auth
//	@Success	307
//	@Failure	307
//	@Router		/api/auth/connect/google/callback [get]
func (a *AuthModule) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	fr := a.frontendURL

	st, err := parseState(state)
	if err != nil {
		st = map[string]string{}
	}

	res, err := a.service.Login(r.Context(), command.Login{
		Code:  code,
		State: st,
	})
	if err != nil {
		slctx.Logger(r.Context()).Error("cannot login", slog.Any("error", err))
		pkg.RedirectError(w, fr, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken.Token,
		Path:     "/",
		Expires:  res.RefreshToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    res.AccessToken.Token,
		Path:     "/",
		Expires:  res.AccessToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, fr.String(), http.StatusPermanentRedirect)
}

// Refresh godoc
//
//	@Summary	generate new jwt pair
//	@Tags		auth
//	@Param		Cookie	header	string	true	"Session cookie"	example("eyJhbGciOiJIUzUx...")
//	@Success	200
//	@Header		200	{string}	Set-Cookie	"Set refresh token cookie; e.g. refresh_token=<token>; HttpOnly; Path=/api/auth/refresh; Secure"
//	@Header		200	{string}	Set-Cookie	"Set access token cookie; e.g. access_token=<token>; HttpOnly; Path=/; Secure"
//	@Failure	400	{object}	response.Error
//	@Failure	401	{object}	response.Error
//	@Failure	500	{object}	response.Error
//	@Router		/api/auth/refresh [get]
func (a *AuthModule) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			slctx.Logger(r.Context()).Debug("refresh token required")
			pkg.RespondError(w, http.StatusBadRequest, response.Error{
				Code:    http.StatusBadRequest,
				Message: "Refresh token required",
			})

			return
		}
		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	res, err := a.service.Refresh(r.Context(), command.Refresh{RefreshToken: c.Value})
	if err != nil {
		slctx.Logger(r.Context()).Error("cannot refresh", slog.Any("error", err))

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken.Token,
		Path:     "/",
		Expires:  res.RefreshToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    res.AccessToken.Token,
		Path:     "/",
		Expires:  res.AccessToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Logout godoc
//
//	@Summary	Logout from session
//	@Tags		auth
//	@Success	200
//	@Failure	400	{object}	response.Error
//	@Failure	401	{object}	response.Error
//	@Router		/api/auth/logout [get]
func (a *AuthModule) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		slctx.Logger(r.Context()).Error("user id not provided in context")
		pkg.RespondError(w, http.StatusBadRequest, response.Error{
			Code:    http.StatusBadRequest,
			Message: "Token not provided",
		})
		return
	}

	sessionID, ok := r.Context().Value(models.SessionIDCtxKey{}).(uuid.UUID)
	if !ok {
		slctx.Logger(r.Context()).Error("session id not provided in context")
		pkg.RespondError(w, http.StatusBadRequest, response.Error{
			Code:    http.StatusBadRequest,
			Message: "Token not provided",
		})
	}

	err := a.service.LogOut(r.Context(), command.Logout{
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		slctx.Logger(r.Context()).Debug("cannot logout", slog.Any("error", err))

		var notFound *domain.ErrorNotFound
		if errors.As(err, &notFound) {
			pkg.RespondError(w, http.StatusNotFound, response.Error{
				Code:    http.StatusNotFound,
				Message: notFound.Message,
			})
		}

		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})

		return
	}

	pkg.RespondOK(w, nil)
}

// UserInfo godoc
//
//	@Summary	Return user info
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	response.UserInfo
//	@Failure	404	{object}	response.Error
//	@Router		/api/auth/user-info [get]
func (a *AuthModule) UserInfo(w http.ResponseWriter, r *http.Request) {
	userInfo, err := a.service.UserInfo(r.Context())
	if err != nil {
		slctx.Logger(r.Context()).Debug("cannot get user info", slog.Any("error", err))
		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	pkg.RespondOK(w, response.UserInfo{
		ID:        userInfo.ID,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		AvatarURL: userInfo.AvatarURL,
		Plan:      userInfo.Plan.Name,
		Theme:     userInfo.Theme,
		CreatedAt: userInfo.CreatedAt,
		UpdatedAt: userInfo.UpdatedAt,
	})
}

func parseState(state string) (map[string]string, error) {
	stateMp := make(map[string]string)
	stateKv := strings.Split(state, "=")

	if len(stateKv)%2 != 0 {
		return nil, fmt.Errorf("invalid state format")
	}

	for i := 0; i < len(stateKv); i += 2 {
		stateMp[stateKv[i]] = stateKv[i+1]
	}

	return stateMp, nil
}
