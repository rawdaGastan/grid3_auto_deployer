// Package validators for validations
package validators

import (
	"errors"
	"net/mail"
	"reflect"

	"github.com/caitlin615/nist-password-validator/password"
	"golang.org/x/crypto/ssh"
)

// ValidateSSHKey used for validating ssh keys
func ValidateSSHKey(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("ValidateSSHKey only validates strings")
	}
	return ValidateSSH(st.String())
}

// ValidateMail used for validating syntax mails
func ValidateMail(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("ValidateMail only validates strings")
	}
	return ValidMail(st.String())
}

// ValidatePassword used for validating passwords before creating user
func ValidatePassword(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("ValidatePassword only validates strings")
	}
	return ValidatePass(st.String())
}

// ValidMail used for validating syntax mails
func ValidMail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// ValidatePass used for validating passwords before creating user
func ValidatePass(pass string) error {
	// password should be ASCII , min 5 , max 10
	validator := password.NewValidator(true, 7, 12)
	err := validator.ValidatePassword(pass)
	if err != nil {
		return err
	}
	return nil
}

// ValidateSSH used for validating ssh keys
func ValidateSSH(sshKey string) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sshKey))
	return err
}
