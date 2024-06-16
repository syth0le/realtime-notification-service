package application

import (
	xservers "github.com/syth0le/gopnik/servers"

	"github.com/go-chi/chi/v5"

	"github.com/syth0le/realtime-notification-service/internal/handler/publicapi"
)

func (a *App) newHTTPServer(env *env) *xservers.HTTPServerWrapper {
	return xservers.NewHTTPServerWrapper(
		a.Logger,
		xservers.WithAdminServer(a.Config.AdminServer),
		xservers.WithPublicServer(a.Config.PublicServer, a.publicMux(env)),
	)
}

func (a *App) publicMux(env *env) *chi.Mux {
	mux := chi.NewMux()

	handler := publicapi.NewHandler(a.Logger, env.notifications)

	mux.Route("/post", func(r chi.Router) {
		r.Use(env.authClient.AuthenticationInterceptor)
		r.HandleFunc("/feed/posted", handler.SubscribeFeedNotifications)
	})

	return mux
}
