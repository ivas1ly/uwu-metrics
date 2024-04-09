package rsakeys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"go.uber.org/zap"
)

func PublicKey(filePath string) (*rsa.PublicKey, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		zap.L().Error("can't open file", zap.Error(err), zap.String("file path", filePath))
		return nil, err
	}

	pemBlock, _ := pem.Decode(file)
	if pemBlock == nil || pemBlock.Type != "RSA PUBLIC KEY" {
		zap.L().Error("failed to decode PEM block containing RSA PUBLIC KEY", zap.String("type", pemBlock.Type))
		return nil, err
	}

	publicKey, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	if err != nil {
		zap.L().Error("can't parse public key", zap.Error(err))
		return nil, err
	}

	return publicKey, nil
}

func PrivateKey(filePath string) (*rsa.PrivateKey, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		zap.L().Error("can't open file", zap.Error(err), zap.String("file path", filePath))
		return nil, err
	}

	pemBlock, _ := pem.Decode(file)
	if pemBlock == nil || pemBlock.Type != "RSA PRIVATE KEY" {
		zap.L().Error("failed to decode PEM block containing RSA PRIVATE KEY", zap.String("type", pemBlock.Type))
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		zap.L().Error("can't parse private key", zap.Error(err))
		return nil, err
	}

	return privateKey, nil
}
