package handlers

import "net/http"

func (app *Application) Logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.InfoLog.Printf("%s %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.RequestURI)

		h.ServeHTTP(w, r)
	})
}

func (app *Application) AdminOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated := app.Sessions.Exists(r.Context(), "authenticated")

		if !authenticated {
			app.clientError(w, http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}
