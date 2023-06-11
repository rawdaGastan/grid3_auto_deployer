// Package internal for internal details
package internal

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/codescalers/cloud4students/validators"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	//go:embed templates/signup.html
	signUpMail []byte

	//go:embed templates/welcome.html
	welcomeMail []byte

	//go:embed templates/reset_pass.html
	resetPassMail []byte

	//go:embed templates/approvedVoucher.html
	approveVoucherMail []byte

	//go:embed templates/rejectedVoucher.html
	rejectedVoucherMail []byte

	//go:embed templates/voucherNotification.html
	notifyVoucherMail []byte

	//go:embed templates/balanceNotification.html
	balanceMail []byte
)

// SendMail sends verification mails
func SendMail(sender, sendGridKey, receiver, subject, body string) error {
	from := mail.NewEmail("Cloud4Students", sender)

	err := validators.ValidMail(receiver)
	if err != nil {
		return fmt.Errorf("email %v is not valid", receiver)
	}

	to := mail.NewEmail("Cloud4Students User", receiver)

	message := mail.NewSingleEmail(from, subject, to, "", body)
	client := sendgrid.NewSendClient(sendGridKey)
	_, err = client.Send(message)

	return err
}

// SignUpMailContent gets the email content for sign up
func SignUpMailContent(code int, timeout int, name string) (string, string) {
	subject := "Welcome to Cloud4Students 🎉"
	body := string(signUpMail)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(name))

	return subject, body
}

// WelcomeMailContent gets the email content for sign up
func WelcomeMailContent(name string) (string, string) {
	subject := "Welcome to Cloud4Students 🎉"
	body := string(welcomeMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(name))

	return subject, body
}

// ResetPasswordMailContent gets the email content for reset password
func ResetPasswordMailContent(code int, timeout int, name string) (string, string) {
	subject := "Reset password"
	body := string(resetPassMail)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(name))

	return subject, body
}

// ApprovedVoucherMailContent gets the content for approved voucher
func ApprovedVoucherMailContent(voucher string, user string) (string, string) {
	subject := "Your voucher request is approved 🎆"
	body := string(approveVoucherMail)

	body = strings.ReplaceAll(body, "-voucher-", fmt.Sprint(voucher))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(user))

	return subject, body
}

// RejectedVoucherMailContent gets the content for rejected voucher
func RejectedVoucherMailContent(user string) (string, string) {
	subject := "Your voucher request is rejected 😔"
	body := string(rejectedVoucherMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(user))

	return subject, body
}

// NotifyAdminsMailContent gets the content for notifying admins
func NotifyAdminsMailContent(vouchers int) (string, string) {
	subject := "There're pending voucher requests for you to review"
	body := string(notifyVoucherMail)

	body = strings.ReplaceAll(body, "-vouchers-", fmt.Sprint(vouchers))

	return subject, body
}

// NotifyAdminsMailLowBalanceContent gets the content for notifying admins when balance is low
func NotifyAdminsMailLowBalanceContent(balance float64) (string, string) {
	subject := "Your account balance is low"
	body := string(balanceMail)

	body = strings.ReplaceAll(body, "-balance-", fmt.Sprint(balance))

	return subject, body
}
