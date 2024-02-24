// Package mailer declares an embedded file system to store our email templates,
// which are stored in the "./templates" directory in the same package.
package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

// Type Mailer is a struct containing a mail.Dialer instance (to connect to an
// SMTP server) and sender information for use in sent emails.
//
// The sender field should be a string of the format "Name <email>".
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New returns an instance of a Mailer struct with the provided SMTP server
// settings. The dialer is configured to have a 5-second timeout when an email
// is sent.
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// The Send method uses the calling Mailer to send an email to the provided
// recipient. Errors are returned if the template file, or its "subject"
// sub-template, can't be parsed. The data object is used to provide data for
// interpolation in the templates.
func (m Mailer) Send(recipient, tmplFile string, data any) error {
	// Parse the provided template file.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+tmplFile)
	if err != nil {
		return err
	}

	// Execute the "plainbody" template from the provided template file, passing
	// in the dynamic data argument, and storing the result in a bytes.Buffer.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the "plainBody" template from the provided template file, passing
	// in the dynamic data argument, and storing the result in a bytes.Buffer.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute the "htmlBody" template from the provided template file, passing
	// in the dynamic data argument, and storing the result in a bytes.Buffer.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Create a new mail.Message instance, setting its To, From, and Subject
	// headers, and setting its body to the template's plain-text body. We also
	// set the HTML body as an alternative.
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String()) // Must call after SetBody

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
