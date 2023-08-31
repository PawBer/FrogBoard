package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"runtime/debug"

	"github.com/PawBer/FrogBoard/internal/models"
)

func (app *Application) populateThreads(threads ...*models.Thread) error {
	for i := 0; i < len(threads); i++ {
		err := app.populateFieldsInPost(&threads[i].Post)
		if err != nil {
			return err
		}

		replies, err := app.ReplyModel.GetLatestReplies(threads[i].BoardID, int(threads[i].ID), 5)
		if err != nil {
			return err
		}

		for j := 0; j < len(replies); j++ {
			err := app.populateFieldsInPost(&replies[j].Post)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (app *Application) populateFieldsInPost(post *models.Post) error {
	files, err := app.FileInfoModel.GetFilesForPost(post.BoardID, post.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	post.Files = files

	citations, err := app.CitationModel.GetCitationsForPost(post.BoardID, post.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	post.Citations = citations

	return nil
}

func (app *Application) createTemplate(requiredTemplates []string) (*template.Template, error) {
	var templateFileNames []string

	requiredFiles, err := fs.ReadDir(app.Templates, "templates/required")
	if err != nil {
		return nil, err
	}

	for _, template := range requiredFiles {
		templateFileNames = append(templateFileNames, fmt.Sprintf("templates/required/%s", template.Name()))
	}

	for _, templateName := range requiredTemplates {
		filePath := fmt.Sprintf("templates/%s.tmpl.html", templateName)
		templateFileNames = append(templateFileNames, filePath)
	}

	tmpl, err := template.ParseFS(app.Templates, templateFileNames...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (app *Application) serverError(w http.ResponseWriter, err error) {
	errorMessage := fmt.Sprintf("%s\n%s\n", err.Error(), debug.Stack())
	app.ErrorLog.Output(2, errorMessage)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
