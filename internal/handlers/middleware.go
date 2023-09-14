package handlers

import (
	"fmt"
	"net/http"
	"time"
)

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

func (app *Application) BlockBannedUsers(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		banned, ban, err := app.BanModel.IsBanned(r)
		if err != nil {
			app.serverError(w, err)
			return
		}

		if banned {
			currentTime := time.Now().UTC()
			if ban.EndDate.Before(currentTime) {
				err := app.BanModel.UnbanUser(ban.IP)
				if err != nil {
					app.serverError(w, err)
					return
				}

				h.ServeHTTP(w, r)
				return
			}

			fmt.Fprintf(w, "You are banned\n")
			fmt.Fprintf(w, "Reason: %s\n", ban.Reason)
			fmt.Fprintf(w, "Ban will end on: %s\n", ban.EndDate)
			return
		}

		h.ServeHTTP(w, r)
	})
}
