package handlers

import (
	"log"
	"net/http"
)

func (app *Application) GetAdmin(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"admin"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	bans, err := app.BanModel.GetBans(0, 15)
	if err != nil {
		app.serverError(w, err)
		return
	}

	users, err := app.UserModel.GetUsers()
	if err != nil {
		app.serverError(w, err)
		return
	}

	latestThreads, err := app.ThreadModel.GetLatestFromEveryBoard()
	if err != nil {
		app.serverError(w, err)
		return
	}

	latestReplies, err := app.ReplyModel.GetLatestRepliesFromEveryThread()
	if err != nil {
		app.serverError(w, err)
		return
	}

	latestFiles, err := app.FileInfoModel.GetLatestFiles()
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
	templateData["Users"] = users
	templateData["LatestThreads"] = latestThreads
	templateData["LatestReplies"] = latestReplies
	templateData["LatestFiles"] = latestFiles

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}
