package gpg

// Package gpg wraps around the gpg command line tool and exposes some of its functionality

import (
	"bytes"
	"os/exec"
)

// Signer implements objects.Signer using gpg
type Signer struct {
	Key string
}

// Sign signs a message b with the key s.Key
func (s Signer) Sign(b []byte) ([]byte, error) {
	cmd := exec.Command("gpg", "--clearsign", "-u", s.Key)

	cmd.Stdin = bytes.NewReader(b)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	return out.Bytes(), err
}

// Verifyer implements objects.Verifyer using gpg
type Verifyer struct{}

// Verify verifies the signed message b
func (Verifyer) Verify(b []byte) error {
	cmd := exec.Command("gpg", "--verify")
	cmd.Stdin = bytes.NewReader(b)
	return cmd.Run()
}
