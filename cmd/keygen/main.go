package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func main() {
	keyPair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Error generating RSA key: %s", err)
		return
	}

	// validate key
	err = keyPair.Validate()
	if err != nil {
		log.Fatalf("Fail to validate key pair, %s", err)
		return
	}

	private := x509.MarshalPKCS1PrivateKey(keyPair)
	if len(private) == 0 {
		log.Fatalf("Error marshalling RSA private key")
		return
	}
	public := x509.MarshalPKCS1PublicKey(&keyPair.PublicKey)
	if len(public) == 0 {
		log.Fatalf("Error marshalling RSA publick key")
		return
	}

	if err := os.WriteFile("private.key", pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: private,
	}), 0777); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("public.key", pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: public,
	}), 0777); err != nil {
		log.Fatal(err)
	}
}
