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
func SendMail(sender string, password string, receiver string, timeout int) (int, error) {
	valid := validator.ValidateMail(receiver)
	if !valid {
		return 0, fmt.Errorf("email %v is not valid", receiver)
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

	subject := "Welcome to Cloud4Students\n\n"
	body := fmt.Sprintf("We are so glad to have you here.\n\nYour code is %s\nThe code will expire in %d minutes.\nPlease don't share it with anyone.", strconv.Itoa(code), timeout)
	message := subject + body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		sender,
		[]string{receiver},
		[]byte(message),
	)
	return code, err
}
