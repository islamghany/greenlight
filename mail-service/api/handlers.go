package api

import (
	"fmt"
	"mailer-service/mailer"
	"net/http"
)

func (server *Server) SendMail(w http.ResponseWriter, r *http.Request) {
	fmt.Print("please work!!")
	var input struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	err := server.readJSON(w, r, &input)
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	data := map[string]interface{}{
		"activationToken": input.Message,
		"userID":          12312,
	}

	msg := mailer.Message{
		From:         input.From,
		To:           input.To,
		Data:         data,
		TemplateFile: "user_welcome.tmpl",
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
