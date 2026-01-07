package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// generateHostKey creates a new Ed25519 SSH host key and saves it to the specified path
func generateHostKey(path string) (ssh.Signer, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Marshal private key to PEM format
	pemBlock, err := gossh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, err
	}

	pemData := pem.EncodeToMemory(pemBlock)
	if err := os.WriteFile(path, pemData, 0600); err != nil {
		return nil, err
	}

	return gossh.NewSignerFromKey(privateKey)
}
