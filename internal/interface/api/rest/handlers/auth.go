package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/auth"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/pkg"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/go-chi/chi/v5"
)

type AuthModule struct {
	service     auth.Service
	frontendURL *url.URL
}

func NewAuthModule(cfg *config.Config, service auth.Service) *AuthModule {
	u, err := url.Parse(cfg.HTTP.FrontendURL)
	if err != nil {
		panic("invalid frontend URL " + err.Error())
	}

	return &AuthModule{
		service:     service,
		frontendURL: u,
	}
}

func (a *AuthModule) Register(r chi.Router) {
	//r.Route("/auth", func(r chi.Router) {
	r.Route("/connect", func(r chi.Router) {
		r.Get("/google", a.Login)
		r.Get("/google/callback", a.Callback)
	})

	r.Get("/refresh", a.Refresh)
	//})
}

// Login godoc
//
//	@Summary	redirect to google oAuth URL
//	@Tags		auth
//	@Param		User-Agent	header	string	true	"<device-name>"	example("PostmanRuntime/7.48.0")
//	@Success	307
//	@Header		307	{string}	Set-Cookie	"Set refresh token cookie; e.g. refresh_token=<token>; HttpOnly; Path=/api/refresh; Secure"
//	@Router		/api/connect/google [get]
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
//	@Router		/api/connect/google/callback [get]
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
		Path:     "/api/refresh",
		Expires:  res.RefreshToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	q := fr.Query()
	q.Set("code", strconv.Itoa(http.StatusOK))
	fr.RawQuery = q.Encode()

	http.Redirect(w, r, fr.String(), http.StatusPermanentRedirect)
}

// Refresh godoc
//
//	@Summary	generate new jwt pair
//	@Tags		auth
//	@Param		Cookie	header		string	true	"Session cookie"	example("eyJhbGciOiJIUzUx...")
//	@Success	200		{object}	response.Refresh
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	500		{object}	response.Error
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
		Path:     "/api/auth/refresh",
		Expires:  res.RefreshToken.ExpireAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	pkg.RespondOK(w, response.Refresh{AccessToken: res.AccessToken.Token})
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
