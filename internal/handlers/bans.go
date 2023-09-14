package handlers

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (app *Application) GetBans(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"bans"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	banCount, err := app.ThreadModel.GetBanCount()
	if err != nil {
		app.serverError(w, err)
		return
	}

	var pageNumber uint

	if banCount <= 30 {
		pageNumber = 0
	} else if r.URL.Query().Has("page") {
		queryPageNumber, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		pageNumber = uint(queryPageNumber) - 1
	} else {
		pageNumber = 0
	}

	pageCount := math.Ceil(float64(banCount) / 30)
	var pageNumbers []int
	for i := 1; i <= int(pageCount); i++ {
		pageNumbers = append(pageNumbers, i)
	}

	bans, err := app.BanModel.GetBans(pageNumber, 30)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["Bans"] = bans
	templateData["PageNumbers"] = pageNumbers

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) GetBanCreate(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"bancreate"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if r.URL.Query().Has("ip") {
		templateData["FormIP"] = r.URL.Query().Get("ip")
	}

	if app.Sessions.Exists(r.Context(), "form-ip") && app.Sessions.Exists(r.Context(), "form-reason") && app.Sessions.Exists(r.Context(), "form-enddate") {
		templateData["FormIP"] = app.Sessions.PopString(r.Context(), "form-ip")
		templateData["FormReason"] = app.Sessions.PopString(r.Context(), "form-reason")
		templateData["FormEndDate"] = app.Sessions.PopString(r.Context(), "form-enddate")
	}

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostBanCreate(w http.ResponseWriter, r *http.Request) {
	formModel := struct {
		IP      string `form:"ip"`
		Reason  string `form:"reason"`
		EndDate string `form:"end-date"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	endDate, err := time.Parse("2006-01-02T15:04", formModel.EndDate)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.BanModel.BanUser(net.ParseIP(formModel.IP), endDate, formModel.Reason)
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while banning the user")

		app.Sessions.Put(r.Context(), "form-ip", formModel.IP)
		app.Sessions.Put(r.Context(), "form-reason", formModel.Reason)
		app.Sessions.Put(r.Context(), "form-enddate", formModel.EndDate)

		http.Redirect(w, r, "/admin/bans/create/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "User banned succesfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *Application) GetBanDelete(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"bandelete"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	ip := chi.URLParam(r, "ip")

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	ban, err := app.BanModel.GetBan(net.ParseIP(ip))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["Ban"] = ban

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostBanDelete(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")

	err := app.BanModel.UnbanUser(net.ParseIP(ip))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "User unbanned successfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
