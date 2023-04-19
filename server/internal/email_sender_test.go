package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendMail(t *testing.T) {
	t.Run("send valid mail", func(t *testing.T) {
		err := SendMail("sender@gmail.com", "1234", "receiver@gmail.com", "subject", "body")
		assert.NoError(t, err)
	})

	t.Run("send invalid mail", func(t *testing.T) {
		err := SendMail("sender@gmail.com", "1234", "receiver", "subject", "body")
		assert.Error(t, err)
	})

}

func TestSignUpMailContent(t *testing.T) {
	subject, body := SignUpMailContent(1234, 60)
	assert.Equal(t, subject, "Welcome to Cloud4Students 🎉")
	assert.Equal(t, body, "We are so glad to have you here.\n\nYour code is 1234\nThe code will expire in 60 seconds.\nPlease don't share it with anyone.")
}

func TestApprovedVoucherMailContent(t *testing.T) {
	subject, body := ApprovedVoucherMailContent("1234", "user")
	assert.Equal(t, subject, "Your voucher is approved 🎆")
	assert.Equal(t, body, "Welcome user,\n\nWe are so glad to inform you that your voucher has been approved successfully.\n\nYour voucher is 1234\n\nBest regards,\nCodescalers team")

}

func TestRejectedVoucherMailContent(t *testing.T) {
	subject, body := RejectedVoucherMailContent("user")
	assert.Equal(t, subject, "Your voucher is rejected 😔")
	assert.Equal(t, body, "Welcome user,\n\nWe are sorry to inform you that your voucher has been rejected\n\nBest regards,\nCodescalers team")

}
