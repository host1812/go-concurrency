package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

// listens for messages
func (m *Mail) sendMail(msg Message, errorChan chan error) {
	app.InfoLog.Println("sendMail: started")
	if msg.Template == "" {
		msg.Template = "mail"
	}
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.From = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	// build html version
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		app.ErrorLog.Println("the was an error creating formatted message, err: ", err)
		errorChan <- err
	}
	app.InfoLog.Println("sendMail: formatted message completed")

	// build text plain version
	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errorChan <- err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Password
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	app.InfoLog.Println("connecting to smtp server")
	smtpClient, err := server.Connect()
	if err != nil {
		log.Println("there was an error connecting to smtp server")
		errorChan <- err
	}
	app.InfoLog.Println("connection established")

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}
	app.InfoLog.Println("sending email")
	err = email.Send(smtpClient)
	if err != nil {
		app.ErrorLog.Println("failed to send email, err:", err)
		app.ErrorLog.Println("email:", email)
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", msg.Template)
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		app.ErrorLog.Println("error executing template, err:", err)
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", msg.Template)
	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()
	if err != nil {
		return "", err
	}

	return plainMessage, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) inlineCSS(s string) (string, error) {
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
