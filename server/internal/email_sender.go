// Package internal for internal details
package internal

import (
	"fmt"

	"strconv"

	"github.com/rawdaGastan/cloud4students/validator"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail sends verification mails
func SendMail(sender, sendGridKey, receiver, subject, body string) error {
	from := mail.NewEmail("Cloud4Students", sender)
	err := validator.ValidateMail(receiver)
	if err != nil {
		return fmt.Errorf("email %v is not valid", receiver)
	}
	to := mail.NewEmail("Cloud4Students User", receiver)

	message := mail.NewSingleEmail(from, subject, to, body, "")

	client := sendgrid.NewSendClient(sendGridKey)
	response, err := client.Send(message)

	log.Debug().Msgf("response: %+v", response)

	return err
}

// SignUpMailContent gets the email content for signup
func SignUpMailContent(code int, timeout int) (string, string) {
	subject := "Welcome to Cloud4Students\n\n"
	body := fmt.Sprintf("We are so glad to have you here.\n\nYour code is %s\nThe code will expire in %d seconds.\nPlease don't share it with anyone.", strconv.Itoa(code), timeout)

	return subject, body
}

// ApprovedVoucherMailContent gets the content for approved voucher
func ApprovedVoucherMailContent(voucher string, user string) (string, string) {
	subject := fmt.Sprintf("Welcome %v,\n\n", user)
	body := fmt.Sprintf("We are so glad to inform you that your voucher has been approved successfully.\n\nYour voucher is %s\n\nBest regards,\nCodescalers team", voucher)

	return subject, body
}
