// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"main/src/lib"
)

// === Read an x509 private key from disk ======================================

func ReadPublicKey(
	publicKeyPath string,
) rsa.PublicKey {

	publicKeyPEM := lib.Read(fmt.Sprintf("%s.pub", publicKeyPath))

	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	if publicKeyBlock == nil {
		lib.Fatal("pem.Decode() failed")
	}
	if publicKeyBlock.Type != "PUBLIC KEY" {
		lib.Fatal("Block is not of type PUBLIC KEY: %v", publicKeyBlock.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		lib.Fatal("x509.ParsePKIXPublicKey() failed: %v", err)
	}

	// Retrieve EK Pub as *rsa.PublicKey
	// See https://stackoverflow.com/a/44317246
	switch ekPubTyp := pubKey.(type) {
	case *rsa.PublicKey:
	default:
		lib.Fatal("ekPublicKey is not of type RSA: %v", ekPubTyp)
	}
	publicKey, _ := pubKey.(*rsa.PublicKey)
	lib.Verbose("publicKey %v", publicKey)

	return *publicKey
}

func ReadPublicKeyPEM(
	publicKeyPEM []byte,
) (
	fatal bool,
	message string,
	key rsa.PublicKey,
) {
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	if publicKeyBlock == nil {
		message = fmt.Sprintf("pem.Decode() failed")
		lib.Error.Print(message)
		return true, message, key
	}
	if publicKeyBlock.Type != "PUBLIC KEY" {
		message = fmt.Sprintf("Block is not of type PUBLIC KEY: %v", publicKeyBlock.Type)
		lib.Error.Print(message)
		return true, message, key
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		message = fmt.Sprintf("x509.ParsePKIXPublicKey() failed: %v", err)
		lib.Error.Print(message)
		return true, message, key
	}

	// Retrieve EK Pub as *rsa.PublicKey
	// See https://stackoverflow.com/a/44317246
	switch ekPubTyp := pubKey.(type) {
	case *rsa.PublicKey:
	default:
		message = fmt.Sprintf("ekPublicKey is not of type RSA: %v", ekPubTyp)
		lib.Error.Print(message)
		return true, message, key
	}
	publicKey, _ := pubKey.(*rsa.PublicKey)
	lib.Verbose("publicKey %v", publicKey)

	return false, "", *publicKey
}
