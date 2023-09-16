package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
)

func (app *Application) GetLogin(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"login"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-username") && app.Sessions.Exists(r.Context(), "form-password") {
		templateData["FormUsername"] = app.Sessions.PopString(r.Context(), "form-username")
		templateData["FormPassword"] = app.Sessions.PopString(r.Context(), "form-password")
	}

	tmpl = tmpl.Funcs(app.getFuncs(r))

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostLogin(w http.ResponseWriter, r *http.Request) {
	formModel := struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user, err := app.UserModel.Login(formModel.Username, formModel.Password)
	if err != nil && errors.Is(err, models.WrongPasswordError{}) {
		app.Sessions.Put(r.Context(), "flash", "The account doesn't exist or password is wrong")

		app.Sessions.Put(r.Context(), "form-username", formModel.Username)
		app.Sessions.Put(r.Context(), "form-password", formModel.Password)

		http.Redirect(w, r, "/login/", http.StatusSeeOther)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.Sessions.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Put(r.Context(), "authenticated", true)
	app.Sessions.Put(r.Context(), "username", user.Username)
	app.Sessions.Put(r.Context(), "display-name", user.DisplayName)
	app.Sessions.Put(r.Context(), "permission", int(user.Permission))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) PostLogout(w http.ResponseWriter, r *http.Request) {
	err := app.Sessions.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Remove(r.Context(), "authenticated")
	app.Sessions.Remove(r.Context(), "username")
	app.Sessions.Remove(r.Context(), "display-name")
	app.Sessions.Remove(r.Context(), "permission")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
