package internal

import "net/mail"

func ValidateMail(address string) bool {
	_, err := mail.ParseAddress(address)
	return err == nil
}
