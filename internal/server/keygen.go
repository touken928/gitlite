package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func generateHostKey(path string) (ssh.Signer, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// 编码为 PEM
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
