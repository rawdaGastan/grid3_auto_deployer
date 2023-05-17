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
	/*
		I really don't think this is the right way to use the validator. If you gonna call
		a validator on specific value with a function then why bother defining custom validators
		with the validator package.

		according to the docs, the validator package is mainly used to add validation annotation on structures
		so for example
		type Config struct{
			Email string `validate:"email"`
		}

		then when let's say u have
		config := //load config

		then do
		validator.Validate(config)

		this will then run validation on all the struct fields based on given config given that
		there is a validator registered with name `email`

		If this is not how u using validator then it's an over kill to just call the ValidateMail function
		directly this way (which has to do reflection!)
	*/
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
	/*
		please use templates instead of this very long and ugly string, then it can also has html and css

		templates can be defined as text files, and then loaded using `embed` package from std.
		you can then use `text/template` or even `html/template` to build a decent template for your email

		this comment applies to all other functions
	*/
	body := fmt.Sprintf("We are so glad to have you here.\n\nYour code is %s\nThe code will expire in %d seconds.\nPlease don't share it with anyone.", strconv.Itoa(code), timeout)

	return subject, body
}

// ApprovedVoucherMailContent gets the content for approved voucher
func ApprovedVoucherMailContent(voucher string, user string) (string, string) {
	subject := "Your voucher request is approved 🎆"
	body := fmt.Sprintf("Welcome %v,\n\nWe are so glad to inform you that your voucher has been approved successfully.\n\nYour voucher is %s\n\nBest regards,\nCodescalers team", user, voucher)

	return subject, body
}

// RejectedVoucherMailContent gets the content for rejected voucher
func RejectedVoucherMailContent(user string) (string, string) {
	subject := "Your voucher request is rejected 😔"
	body := fmt.Sprintf("Welcome %v,\n\nWe are sorry to inform you that your voucher request has been rejected.\n\nBest regards,\nCodescalers team", user)

	return subject, body
}

// NotifyAdminsMailContent gets the content for notifying admins
func NotifyAdminsMailContent(vouchers int) (string, string) {
	subject := "There're pending voucher requests for you to review"
	body := fmt.Sprintf("Hello,\n\nThere are %d voucher requests that need to be reviewed. Kindly check them.\n\nBest regards,\nCodescalers team", vouchers)

	return subject, body
}
