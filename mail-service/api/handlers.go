package api

import (
	"mailer-service/mailer"
	"net/http"
)

func (server *Server) SendMailHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From         string                 `json:"from"`
		To           string                 `json:"to"`
		Data         map[string]interface{} `json:"data"`
		TemplateFile string                 `json:"template_file"`
		Attachment   []string               `json:"attachments"`
	}

	err := server.readJSON(w, r, &input, true)
	if err != nil {
		server.badRequestResponse(w, r, err)
		return
	}

	if input.TemplateFile == "" {
		server.badRequestResponse(w, r, mailer.ErrTemplateNotFound)
		return
	}
	msg := mailer.Message{
		From:         input.From,
		To:           input.To,
		Data:         input.Data,
		TemplateFile: input.TemplateFile,
		Attachments:  input.Attachment,
	}

	err = server.mailer.Send(msg)

	if err != nil {
		if err == mailer.ErrTemplateNotFound {
			server.badRequestResponse(w, r, err)

		} else {
			server.serverErrorResponse(w, r, err)
		}
		return
	}

	err = server.writeJson(w, http.StatusCreated, envelope{"message": "email sent"}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}
}
