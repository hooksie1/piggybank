package service

import (
	"testing"
)

func TestToBase64(t *testing.T) {

	tests := []struct {
		expected string
		word     string
	}{
		{"dGVzdGluZw", "testing"},
		{"dGVzdA", "test"},
		{"cGlnZ3ktYmFuaw", "piggy-bank"},
	}

	for _, tt := range tests {
		decoded := []byte(tt.word)

		encoded := toBase64(decoded)

		if encoded != tt.expected {
			t.Errorf("Expected %s, but got %s", tt.expected, encoded)
		}
	}

}

func TestFromBase64(t *testing.T) {
	tests := []struct {
		encoded  string
		expected string
	}{
		{"dGVzdGluZw", "testing"},
		{"dGVzdA", "test"},
		{"cGlnZ3ktYmFuaw", "piggy-bank"},
	}

	for _, tt := range tests {
		encoded := tt.encoded

		decoded, err := fromBase64(encoded)
		if err != nil {
			t.Fatal(err)
		}

		if string(decoded) != tt.expected {
			t.Errorf("Expected %s, but got %s", tt.expected, decoded)
		}
	}
}
