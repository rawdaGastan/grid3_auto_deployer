package internal

import (
	"testing"
)

func TestReadConfFile(t *testing.T) {
	data, err := ReadConfFile("/home/alaa/codescalers/cloud4students/server/config.json") //TODO: Is it right to put path like this??
	if err != nil {
		t.Error(err)
	}
	if data == nil {
		t.Errorf("File is empty!")
	}
}

func TestParseConf(t *testing.T) {
	data, err := ReadConfFile("/home/alaa/codescalers/cloud4students/server/config.json")
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
	if got.MailSender != expected.MailSender {
		t.Errorf("incorrect data, got %v, want %v", got.MailSender, expected.MailSender)
	}
	if got.Account.Mnemonics != expected.Account.Mnemonics {
		t.Errorf("incorrect data, got %s, want %s", got.Account.Mnemonics, expected.Account.Mnemonics)
	}

}
