package internal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestSendMail(t *testing.T) {
	m := NewMailer("1234")

	t.Run("send valid mail", func(t *testing.T) {
		err := m.SendMail("sender@gmail.com", "receiver@gmail.com", "subject", "body")
		assert.NoError(t, err)
	})

	t.Run("send invalid mail", func(t *testing.T) {
		err := m.SendMail("sender@gmail.com", "receiver", "subject", "body")
		assert.Error(t, err)
	})
}

func TestSignUpMailContent(t *testing.T) {
	subject, body := SignUpMailContent(1234, 60, "user", "")
	assert.Equal(t, subject, "Welcome to Cloud4All 🎉")

	want := string(signUpMail)
	want = strings.ReplaceAll(want, "-code-", fmt.Sprint(1234))
	want = strings.ReplaceAll(want, "-time-", fmt.Sprint(60))
	want = strings.ReplaceAll(want, "-name-", cases.Title(language.Und).String("user"))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestResetPassMailContent(t *testing.T) {
	subject, body := ResetPasswordMailContent(1234, 60, "user", "")
	assert.Equal(t, subject, "Reset password")

	want := string(resetPassMail)
	want = strings.ReplaceAll(want, "-code-", fmt.Sprint(1234))
	want = strings.ReplaceAll(want, "-time-", fmt.Sprint(60))
	want = strings.ReplaceAll(want, "-name-", cases.Title(language.Und).String("user"))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestApprovedVoucherMailContent(t *testing.T) {
	subject, body := ApprovedVoucherMailContent("1234", "user", "")
	assert.Equal(t, subject, "Your voucher request is approved 🎆")

	want := string(approveVoucherMail)
	want = strings.ReplaceAll(want, "-voucher-", fmt.Sprint(1234))
	want = strings.ReplaceAll(want, "-name-", cases.Title(language.Und).String("user"))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestRejectedVoucherMailContent(t *testing.T) {
	subject, body := RejectedVoucherMailContent("user", "")
	assert.Equal(t, subject, "Your voucher request is rejected 😔")

	want := string(rejectedVoucherMail)
	want = strings.ReplaceAll(want, "-name-", cases.Title(language.Und).String("user"))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestNotifyVoucherMailContent(t *testing.T) {
	subject, body := NotifyAdminsMailContent(7, "")
	assert.Equal(t, subject, "There're pending voucher requests for you to review")

	want := string(notifyVoucherMail)
	want = strings.ReplaceAll(want, "-vouchers-", fmt.Sprint(7))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestNotifyBalanceMailContent(t *testing.T) {
	subject, body := NotifyAdminsMailLowBalanceContent(200, "")
	assert.Equal(t, subject, "Your account balance is low")

	want := string(balanceMail)
	want = strings.ReplaceAll(want, "-balance-", fmt.Sprint(200))
	want = strings.ReplaceAll(want, "-host-", "")

	assert.Equal(t, body, want)
}

func TestAdminAnnouncementMailContent(t *testing.T) {
	subject, body := AdminAnnouncementMailContent("subject!", "announcement!", "", "")
	assert.Equal(t, subject, "New Announcement! 📢 subject!")
	want := string(adminAnnouncement)
	want = strings.ReplaceAll(want, "-subject-", "subject!")
	want = strings.ReplaceAll(want, "-body-", "announcement!")
	want = strings.ReplaceAll(want, "-host-", "")
	want = strings.ReplaceAll(want, "-name-", "")
	assert.Equal(t, body, want)
}
