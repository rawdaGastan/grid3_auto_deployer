// Package internal for internal details
package internal

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/rawdaGastan/cloud4students/validator"
)

// SendMail sends verification mails
func SendMail(sender string, password string, receiver string, message string) error {
	err := validator.ValidateMail(receiver)
	if err != nil {
		return fmt.Errorf("email %v is not valid", receiver)
	}
	auth := smtp.PlainAuth(
		"",
		sender,
		password,
		"smtp.gmail.com",
	)

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		sender,
		[]string{receiver},
		[]byte(message),
	)
	return err
}

// SignUpMailBody gets the email body for signup
func SignUpMailBody(code int, timeout int) string {
	subject := "Welcome to Cloud4Students\n\n"
	body := fmt.Sprintf("We are so glad to have you here.\n\nYour code is %s\nThe code will expire in %d seconds.\nPlease don't share it with anyone.", strconv.Itoa(code), timeout)
	message := subject + body

	return message
}

// ApprovedVoucherMailBody gets the body for approved voucher
func ApprovedVoucherMailBody(voucher string, user string) string {
	subject := fmt.Sprintf("Welcome %v,\n\n", user)
	body := fmt.Sprintf("We are so glad to inform you that your voucher has been approved successfully.\n\nYour voucher is %s\n\nBest regards,\nCodescalers team", voucher)
	message := subject + body

	return message
}
