package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
)

const (
	// smaller size causes error "message too long for RSA key size".
	bitSize = 16384
)

//go:generate go run generate.go
func main() {
	localDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("can't get local working directory: %s\n", err.Error())
	}
	fmt.Println("current directory: ", localDir)

	cmdDir := path.Join(localDir, "../../cmd/")

	privateRSAKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		fmt.Printf("can't generate RSA key: %s\n", err.Error())
		return
	}
	generatePrivateRSAKey(privateRSAKey, cmdDir)

	publicRSAKey := &privateRSAKey.PublicKey
	generatePublicRSAKey(publicRSAKey, cmdDir)
}

func generatePrivateRSAKey(key *rsa.PrivateKey, dir string) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privateKeyFile, err := os.Create(path.Join(dir+"/server/", "private_key.pem"))
	if err != nil {
		fmt.Printf("unable to create new file: %s\n", err.Error())
		return
	}

	err = pem.Encode(privateKeyFile, privateKeyBlock)
	if err != nil {
		fmt.Printf("unable to save private key to file: %s", err.Error())
		return
	}

	fmt.Printf("private key saved %s\n", privateKeyFile.Name())

	err = privateKeyFile.Close()
	if err != nil {
		fmt.Printf("can't close file: %s\n", err.Error())
		return
	}
}

func generatePublicRSAKey(key *rsa.PublicKey, dir string) {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(key)
	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicKeyFile, err := os.Create(path.Join(dir+"/agent/", "public_key.pem"))
	if err != nil {
		fmt.Printf("unable to create new file: %s\n", err.Error())
		return
	}

	err = pem.Encode(publicKeyFile, publicKeyBlock)
	if err != nil {
		fmt.Printf("unable to save public key to file: %s\n", err.Error())
		return
	}

	fmt.Printf("public key saved %s\n", publicKeyFile.Name())

	err = publicKeyFile.Close()
	if err != nil {
		fmt.Printf("can't close file: %s\n", err.Error())
		return
	}
}
