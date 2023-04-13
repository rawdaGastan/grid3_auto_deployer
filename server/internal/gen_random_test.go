package internal

import "testing"

func TestGenerateRandomVoucher(t *testing.T) {
	voucher := GenerateRandomVoucher(10)
	if len(voucher) != 10 {
		t.Errorf("Expected voucher length to be 10, got %d", len(voucher))
	}
}

func TestGenerateRandomCode(t *testing.T) {
	code := GenerateRandomCode()
	if code < 1000 || code > 9999 {
		t.Errorf("Expected code to be between 1000 and 9999, got %d", code)
	}
}	