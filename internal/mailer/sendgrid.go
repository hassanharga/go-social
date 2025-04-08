package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	from   string
	apiKey string
	client *sendgrid.Client
}

func NewSendGridMailer(apiKey, from string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		from:   from,
		apiKey: apiKey,
		client: client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromEmail, m.from)
	to := mail.NewEmail(username, email)

	// building and parsing the template
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := range MaxRetries {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Error sending email: %v, attempt %d of %d", email, i+1, MaxRetries)
			log.Printf("Error: %v\n", err)

			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		fmt.Printf("Email sent with status code: %d\n", response.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", MaxRetries)
}
