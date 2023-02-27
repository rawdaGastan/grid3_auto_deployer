package internal

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"time"

	"github.com/rawdaGastan/grid3_auto_deployer/validator"
)

func SendMail(sender string, password string, reciever string, subject string, body string) (int, error) {
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
	code := rand.Intn(max-min) + min //TODO: body in ``
	message := subject + body
	// "Welcome to Cloude4Students,\n" +
	// 	"we are so glad you are here,\n" +
	// 	"Your code is " + strconv.Itoa(code) + "\n" +
	// 	"The code will expire in 5 min\n" +
	// 	"Please don't share the code with anyone"

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		sender,
		[]string{reciever},
		[]byte(message),
	)
	return code, err
}
