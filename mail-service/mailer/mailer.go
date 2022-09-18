package mailer

import (
	"bytes"
	"crypto/tls"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/vanng822/go-premailer/premailer"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓

//go:embed "templates"
var templateFS embed.FS

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From         string
	FromName     string
	To           string
	TemplateFile string
	Attachments  []string
	Data         interface{}
}

type Mailer struct {
	dialer *mail.Dialer
	Mail
}

func NewMailer(m Mail) *Mailer {
	// first we have to initilize the mailer sender

	dialer := mail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	dialer.Timeout = 5 * time.Second

	if m.Encryption == "tls" {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else if m.Encryption == "ssl" {
		dialer.SSL = true
	}

	return &Mailer{
		Mail:   m,
		dialer: dialer,
	}
}

func (m *Mailer) Send(msg Message) error {

	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+msg.TemplateFile)

	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", msg.Data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", msg.Data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", msg.Data)
	if err != nil {
		return err
	}

	formatedMessageBody := htmlBody.String()

	formatedMessageBody, err = m.inlineCSS(formatedMessageBody)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then we use the SetHeader() method to set the email recipient, sender and subject
	// headers, the SetBody() method to set the plain-text body, and the AddAlternative()
	// method to set the HTML body. It's important to note that AddAlternative() should
	// always be called *after* SetBody().
	server := mail.NewMessage()
	server.SetHeader("To", msg.To)
	server.SetHeader("From", msg.From)
	server.SetHeader("Subject", subject.String())
	server.SetBody("text/plain", plainBody.String())
	server.AddAlternative("text/html", htmlBody.String())
	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			server.Attach(x)
		}
	}

	// Call the DialAndSend() method on the dialer, passing in the message to send. This
	// opens a connection to the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
	// error.
	err = m.dialer.DialAndSend(server)
	if err != nil {
		return err
	}

	return nil
}

func (mailer *Mailer) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}
