package api

import (
	"mailer-service/mailer"
	"net/http"
)

func (server *Server) SendMail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From       string   `json:"from"`
		To         string   `json:"to"`
		Subject    string   `json:"subject"`
		Message    string   `json:"message"`
		Attachment []string `json:"attachments"`
	}

	err := server.readJSON(w, r, &input, true)
	if err != nil {
		server.badRequestResponse(w, r, err)
		return
	}

	data := map[string]interface{}{
		"subject": input.Subject,
		"message": input.Message,
	}

	msg := mailer.Message{
		From:         input.From,
		To:           input.To,
		Data:         data,
		TemplateFile: "user_welcome.tmpl",
		Attachments:  input.Attachment,
	}

	err = server.mailer.Send(msg)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	err = server.writeJson(w, http.StatusCreated, envelope{"message": "email sent"}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}
}
