package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"text/template"

	"github.com/shopspring/decimal"
)

type Mail struct {
	Sender  string
	To      string
	Subject string
	Body    bytes.Buffer
}

type User struct {
	Name  string
	Email string
	Debt  decimal.Decimal
}

func main() {

	sender := "john.doe@example.com"

	var users = []User{
		{"Roger Roe", "roger.roe@example.com", decimal.NewFromFloat(890.50)},
		{"Peter Smith", "peter.smith@example.com", decimal.NewFromFloat(350)},
		{"Lucia Green", "lucia.green@example.com", decimal.NewFromFloat(120.80)},
	}

	my_user := "9c1d45eaf7af5b"
	my_password := "ad62926fa75d0f"
	addr := "smtp.mailtrap.io:2525"
	host := "smtp.mailtrap.io"

	subject := "Amount due"

	var template_data = `
    Dear {{ .Name }}, your debt amount is ${{ .Debt }}.`

	for _, user := range users {

		t := template.Must(template.New("template_data").Parse(template_data))
		var body bytes.Buffer

		err := t.Execute(&body, user)
		if err != nil {
			log.Fatal(err)
		}

		request := Mail{
			Sender:  sender,
			To:      user.Email,
			Subject: subject,
			Body:    body,
		}

		msg := BuildMessage(request)
		auth := smtp.PlainAuth("", my_user, my_password, host)
		err2 := smtp.SendMail(addr, auth, sender, []string{user.Email}, []byte(msg))

		if err2 != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Emails sent successfully")
}

func BuildMessage(mail Mail) string {
	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", mail.To)
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body.String())

	return msg
}
