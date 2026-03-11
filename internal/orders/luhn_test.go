package orders

import "testing"

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid order number from spec", "12345678903", true},
		{"valid order number from spec 2", "9278923470", true},
		{"valid single zero", "0", true},
		{"valid credit card number", "4532015112830366", true},
		{"valid with doubling over 9", "79927398713", true},
		{"invalid checksum", "12345678901", false},
		{"invalid single digit", "1", false},
		{"empty string", "", false},
		{"non-digit characters", "1234abc", false},
		{"spaces in number", "1234 5678", false},
		{"single letter", "a", false},
		{"valid two digits", "18", true},
		{"invalid two digits", "19", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateLuhn(tt.number); got != tt.want {
				t.Errorf("ValidateLuhn(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}