package routes

import (
	"math/rand"
	"net/smtp"
	"strconv"
	"time"
)

func SendMail(reciever string) (int, error) {
	auth := smtp.PlainAuth(
		"",
		"alaamahmoud.1223@gmail.com",
		"iqpfshurvllcknpl",
		"smtp.gmail.com",
	)

	// generate random code of 4 digits
	min := 1000
	max := 9999
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(max-min) + min
	message := "Subject: CodeScalers Egypt\n" + strconv.Itoa(code) + "\n" + "The code will expire in 5 min"

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"alaamahmoud.1223@gmail.com",
		[]string{reciever},
		[]byte(message),
	)
	return code, err
}
