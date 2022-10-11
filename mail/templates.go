package mail

import (
	"fmt"
)

type emailTemplate struct {
	Subject string
	Body    string
}

var templates map[string]emailTemplate = make(map[string]emailTemplate)

func initTemplates() {
	templates["login.post_login_request"] = emailTemplate{
		Subject: "Bore app login OTP",
		Body: `
		Hello, {{.user.Email}}!<br />
		Your login OTP is: <b>{{.otp}}</b>
		`,
	}

}
func getTemplate(name string) (*emailTemplate, error) {
	t, ok := templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s not exists", name)
	}

	return &t, nil
}
