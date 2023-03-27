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
	_, err := ssh.ParsePublicKey([]byte(st.String()))
	return err
}

// ValidateMail used for validating syntax mails
func ValidateMail(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("ValidateMail only validates strings")
	}
	_, err := mail.ParseAddress(st.String())
	return err
}

// ValidatePassword used for validating passwords before creating user
func ValidatePassword(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("ValidatePassword only validates strings")
	}
	// password should be ASCII , min 5 , max 10
	validator := password.NewValidator(true, 5, 10)
	err := validator.ValidatePassword(st.String())
	if err != nil {
		return err
	}
	return nil
}
