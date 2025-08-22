package mailer

import (
	"bytes"
	"gopkg.in/gomail.v2"
	"html/template"
)

type Mailer struct {
	host     string
	port     int
	username string
	password string
}

func NewMailer(host string, port int, username string, password string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (m *Mailer) SendMailFromTemplate(toEmail string, subject string, templatePath string, templateData any) error {

	tmpl := template.Must(template.ParseFiles(templatePath))

	var body bytes.Buffer

	if err := tmpl.Execute(&body, templateData); err != nil {
		return err
	}

	message := gomail.NewMessage()

	message.SetHeader("From", m.username)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body.String())

	d := gomail.NewDialer(m.host, m.port, m.username, m.password)

	return d.DialAndSend(message)
}
