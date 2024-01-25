package mail

import (
	"bytes"
	"crypto/tls"
	"text/template"

	"github.com/fullpipe/bore-server/entity"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer    *gomail.Dialer
	templates *template.Template
}

type MailerConfig struct {
	Host     string `required:"true"`
	Port     int    `default:"1025"`
	Username string
	Password string
}

func NewMailer(cfg MailerConfig) (*Mailer, error) {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	mailer := Mailer{dialer: d}
	initTemplates()

	err := mailer.initTemplates()
	if err != nil {
		return nil, err
	}

	return &mailer, nil
}

func (mailer *Mailer) initTemplates() error {
	mailer.templates = template.New("base")

	return nil
}

func WithParam(key string, value any) func(params map[string]any) {
	return func(params map[string]any) {
		params[key] = value
	}
}

func WithParams(in map[string]any) func(params map[string]any) {
	return func(params map[string]any) {
		for key, p := range in {
			params[key] = p
		}
	}
}

func (mailer *Mailer) SendToEmail(message, email string, wps ...func(params map[string]any)) error {
	m := gomail.NewMessage()

	m.SetHeader("From", "noreply@bore.app")
	m.SetHeader("To", email)

	tmp, err := getTemplate(message)
	if err != nil {
		return err
	}

	t, err := template.New("Subject").Parse(tmp.Subject)
	if err != nil {
		return err
	}

	params := make(map[string]any)
	for _, wp := range wps {
		wp(params)
	}

	subject := bytes.NewBufferString("")
	err = t.Execute(subject, params)
	if err != nil {
		return err
	}

	t, err = template.New("Body").Parse(tmp.Body)
	if err != nil {
		return err
	}
	body := bytes.NewBufferString("")
	err = t.Execute(body, params)
	if err != nil {
		return err
	}

	m.SetHeader("Subject", subject.String())
	m.SetBody("text/html", body.String())

	return mailer.dialer.DialAndSend(m)
}

func (mailer *Mailer) SendToUser(message string, user *entity.User, params map[string]any) error {
	return mailer.SendToEmail(message, user.Email, WithParam("user", user))
}
