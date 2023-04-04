// Package internal for internal details
package internal

import (
	"fmt"

	"strconv"

	"github.com/codescalers/cloud4students/validators"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail sends verification mails
func SendMail(sender, sendGridKey, receiver, subject, body string) error {
	from := mail.NewEmail("Cloud4Students", sender)
	err := validators.ValidateMail(receiver, "")
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
	subject := "Welcome to Cloud4Students 🎉"
	body := fmt.Sprintf("We are so glad to have you here.\n\nYour code is %s\nThe code will expire in %d seconds.\nPlease don't share it with anyone.", strconv.Itoa(code), timeout)

	return subject, body
}

// ApprovedVoucherMailContent gets the content for approved voucher
func ApprovedVoucherMailContent(voucher string, user string) (string, string) {
	subject := "Your voucher is approved 🎆"
	body := fmt.Sprintf("Welcome %v,\n\nWe are so glad to inform you that your voucher has been approved successfully.\n\nYour voucher is %s\n\nBest regards,\nCodescalers team", user, voucher)

	return subject, body
}

// RejectedVoucherMailContent gets the content for rejected voucher
func RejectedVoucherMailContent(user string) (string, string) {
	subject := "Your voucher is rejected 😔"
	body := fmt.Sprintf("Welcome %v,\n\nWe are sorry to inform you that your voucher has been rejected\n\nBest regards,\nCodescalers team", user)

	return subject, body
}
