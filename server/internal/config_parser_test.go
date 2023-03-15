package internal

import (
	"testing"
)

func TestReadConfFile(t *testing.T) {
	data, err := ReadConfFile("../tests/config-temp.json")
	if err != nil {
		t.Error(err)
	}
	if data == nil {
		t.Errorf("File is empty!")
	}
}

func TestParseConf(t *testing.T) {
	data, err := ReadConfFile("../tests/config-temp.json")
	if err != nil {
		t.Error(err)
	}
	expected := Configuration{
		Server: Server{
			Host: "localhost",
			Port: ":3000",
		},
		MailSender: MailSender{
			Email:    "alaamahmoud.1223@gmail.com",
			Password: "iqpfshurvllcknpl",
		},
		Account: GridAccount{
			Mnemonics: "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		},
	}
	got, err := ParseConf(data)
	if err != nil {
		t.Error(err)
	}
	if got.Server != expected.Server {
		t.Errorf("incorrect data, got %v, want %v", got.Server, expected.Server)
	}
	if got.MailSender.Email != expected.MailSender.Email {
		t.Errorf("incorrect data, got %v, want %v", got.MailSender.Email, expected.MailSender.Email)
	}
	if got.Account.Mnemonics != expected.Account.Mnemonics {
		t.Errorf("incorrect data, got %s, want %s", got.Account.Mnemonics, expected.Account.Mnemonics)
	}

}
