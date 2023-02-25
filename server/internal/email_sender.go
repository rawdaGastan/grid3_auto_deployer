package internal

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"

	"github.com/rawdaGastan/grid3_auto_deployer/validator"
)

var email string = "alaamahmoud.1223@gmail.com" //TODO: will be changed

func SendMail(reciever string) (int, error) {
	valid := validator.ValidateMail(reciever)
	if !valid {
		return 0, fmt.Errorf("email %v is not valid", reciever)
	}
	auth := smtp.PlainAuth(
		"",
		email,
		"iqpfshurvllcknpl",
		"smtp.gmail.com",
	)

	// generate random code of 4 digits
	min := 1000
	max := 9999
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(max-min) + min
	message := "Subject: Cloud4students \n" +
		"Welcome to Cloude4Students,\n" +
		"we are so glad you are here,\n" +
		"Your code is " + strconv.Itoa(code) + "\n" +
		"The code will expire in 5 min\n" +
		"Please don't share the code with anyone"

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		email,
		[]string{reciever},
		[]byte(message),
	)
	return code, err
}
