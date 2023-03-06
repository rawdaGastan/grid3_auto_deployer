// Package internal for internal details
package internal

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"

	"github.com/rawdaGastan/cloud4students/validator"
)

// SendMail sends verification mails
func SendMail(sender string, password string, reciever string) (int, error) {
	valid := validator.ValidateMail(reciever)
	if !valid {
		return 0, fmt.Errorf("email %v is not valid", reciever)
	}
	auth := smtp.PlainAuth(
		"",
		sender,
		password,
		"smtp.gmail.com",
	)

	// generate random code of 4 digits
	min := 1000
	max := 9999
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(max-min) + min

	subject := "Welcome to Cloud4Students. \n"
	body := `We are so glad to have you here
your code is` + strconv.Itoa(code) +
		`The code will expire in 5 minutes
Please don't share it with anyone.`
	message := subject + body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		sender,
		[]string{reciever},
		[]byte(message),
	)
	return code, err
}
