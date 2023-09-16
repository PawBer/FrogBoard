package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/go-chi/chi/v5"
)

func (app *Application) GetUserCreate(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"usercreate"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-username") && app.Sessions.Exists(r.Context(), "form-displayname") {
		templateData["FormUsername"] = app.Sessions.PopString(r.Context(), "form-username")
		templateData["FormDisplayName"] = app.Sessions.PopString(r.Context(), "form-displayname")
	}

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostUserCreate(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	formModel := struct {
		Username    string `form:"username"`
		DisplayName string `form:"display-name"`
		Permission  string `form:"permission"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	permission, err := strconv.ParseInt(formModel.Permission, 10, 32)
	if err != nil {
		app.serverError(w, err)
		return
	}

	password, err := app.UserModel.RegisterUser(formModel.Username, formModel.DisplayName, models.UserPermission(permission))
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while creating the user")

		app.Sessions.Put(r.Context(), "form-username", formModel.Username)
		app.Sessions.Put(r.Context(), "form-displayname", formModel.DisplayName)

		http.Redirect(w, r, "/admin/users/create/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "User created succesfully")
	app.Sessions.Put(r.Context(), "password", password)

	http.Redirect(w, r, "/admin/users/create/success/", http.StatusSeeOther)
}

func (app *Application) GetUserCreateSuccess(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"usercreatesuccess"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !app.Sessions.Exists(r.Context(), "password") {
		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
	}

	templateData["Password"] = app.Sessions.PopString(r.Context(), "password")

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) GetUserEdit(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"useredit"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	username := chi.URLParam(r, "username")

	user, err := app.UserModel.GetUser(username)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-username") && app.Sessions.Exists(r.Context(), "form-displayname") {
		templateData["FormDisplayName"] = app.Sessions.PopString(r.Context(), "form-displayname")
	}

	templateData["User"] = user

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}

}

func (app *Application) PostUserEdit(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	formModel := struct {
		Username    string `form:"username"`
		DisplayName string `form:"display-name"`
		Permission  string `form:"permission"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	permission, err := strconv.ParseInt(formModel.Permission, 10, 32)
	if err != nil {
		app.serverError(w, err)
		return
	}

	newUser := models.User{
		Username:    formModel.Username,
		DisplayName: formModel.DisplayName,
		Permission:  models.UserPermission(permission),
	}

	err = app.UserModel.Update(newUser)
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while creating the user")

		app.Sessions.Put(r.Context(), "form-displayname", formModel.DisplayName)

		http.Redirect(w, r, "/admin/users/create/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "User edited succesfully")

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *Application) GetUserDelete(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"userdelete"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	username := chi.URLParam(r, "username")

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user, err := app.UserModel.GetUser(username)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["User"] = user

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostUserDelete(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	username := chi.URLParam(r, "username")

	err := app.UserModel.Delete(username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "User deleted successfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *Application) GetPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		fmt.Printf("Not an admin\n")
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"userpasswordreset"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	username := chi.URLParam(r, "username")
	user, err := app.UserModel.GetUser(username)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["User"] = user
	templateData["OwnUsername"] = user.Username == app.Sessions.GetString(r.Context(), "username")

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	username := chi.URLParam(r, "username")
	password, err := app.UserModel.ResetUserPassword(username)
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while resetting the password")

		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "password", password)

	url := fmt.Sprintf("/admin/users/%s/passwordreset/success/", username)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (app *Application) GetPasswordResetSuccess(w http.ResponseWriter, r *http.Request) {
	if !app.hasPermission(r, models.Admin) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	requiredTemplates := []string{"userpasswordresetsuccess"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !app.Sessions.Exists(r.Context(), "password") {
		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
	}

	templateData["Password"] = app.Sessions.PopString(r.Context(), "password")

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) GetPasswordChange(w http.ResponseWriter, r *http.Request) {
	if !app.Sessions.Exists(r.Context(), "authenticated") {
		app.clientError(w, http.StatusForbidden)
	}

	requiredTemplates := []string{"userpasswordchange"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-password") {
		templateData["FormPassword"] = app.Sessions.PopString(r.Context(), "form-password")
	}

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostPasswordChange(w http.ResponseWriter, r *http.Request) {
	if !app.Sessions.Exists(r.Context(), "authenticated") {
		app.clientError(w, http.StatusForbidden)
		return
	}

	formModel := struct {
		Password string `form:"password"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.UserModel.ChangeUserPassword(app.Sessions.GetString(r.Context(), "username"), formModel.Password)
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while changing password")

		app.Sessions.Put(r.Context(), "form-password", formModel.Password)

		http.Redirect(w, r, "/admin/users/passwordchange/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "Password changed successfully")

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
