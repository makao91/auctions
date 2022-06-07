package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"strings"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

func sendEmail() {

	sender := "john.doe@example.com"

	to := []string{
		"roger.roe@example.com",
	}

	user := "91399c960760af"
	password := "fb7986dd7608c4"

	subject := "Simple HTML mail"
	body := `<p>An old <b>falcon</b> in the sky.</p>`

	request := Mail{
		Sender:  sender,
		To:      to,
		Subject: subject,
		Body:    body,
	}

	addr := "smtp.mailtrap.io:2525"
	host := "smtp.mailtrap.io"

	msg := BuildMail(request)
	auth := smtp.PlainAuth("", user, password, host)
	err := smtp.SendMail(addr, auth, sender, to, []byte(msg))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email sent successfully")
}

func BuildMail(mail Mail) []byte {

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", mail.Sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mail.Subject))

	boundary := "my-boundary-779"
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n",
		boundary))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString(fmt.Sprintf("\r\n%s", mail.Body))

	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString("Content-Disposition: attachment; filename=filtered_auction.txt\r\n")
	buf.WriteString("Content-ID: <filtered_auction.txt>\r\n\r\n")

	data := readFile("source/filtered_auction.txt")

	b := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b, data)
	buf.Write(b)
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))

	buf.WriteString("--")

	return buf.Bytes()
}

func readFile(fileName string) []byte {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return data
}
