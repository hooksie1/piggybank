package server

import (
	"testing"
)

type TB struct {
	name     string
	expected string
	actual   string
}

func TestGeneratePass(t *testing.T) {
	secret := generatePass()
	secret2 := generatePass()

	if len(secret) != 43 {
		t.Errorf("Secret is not 32 bytes")
	}

	if secret == secret2 {
		t.Errorf("Secrets are the same.")
	}

}
