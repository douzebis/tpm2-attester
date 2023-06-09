// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"main/src/certs"
	"main/src/lib"

	"github.com/google/go-attestation/attest"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// === Verifier: verify quote ==================================================

func VerifyQuote(
	verifierAkPath string, // IN
	verifierNoncePath string, // IN
	cicdDigestPath string, // IN
	attestorQuotePath string, // IN
) {

	lib.PRINT("=== VERIFIER: VERIFY QUOTE =====================================================")

	// Read nonce, attestation and signature from disk
	nonce := lib.Read(fmt.Sprintf("%s.bin", verifierNoncePath))
	attestation := lib.Read(fmt.Sprintf("%s-attest.bin", attestorQuotePath))
	signature := lib.Read(fmt.Sprintf("%s-signature.bin", attestorQuotePath))

	att, err := tpm2.DecodeAttestationData(attestation)
	if err != nil {
		lib.Fatal("DecodeAttestationData() failed: %v", err)
	}

	lib.Verbose("Attestation ExtraData (nonce): 0x%s ", hex.EncodeToString(att.ExtraData))
	lib.Verbose("Attestation PCR#: %v ", att.AttestedQuoteInfo.PCRSelection.PCRs)
	lib.Verbose("Attestation Hash: 0x%s ", hex.EncodeToString(att.AttestedQuoteInfo.PCRDigest))

	// Compare the nonce that is embedded within the attestation. This should
	// match the one we sent in earlier.
	if !bytes.Equal(nonce, att.ExtraData) {
		lib.Fatal("Nonce Value mismatch Got: (0x%s) Expected: (0x%s)",
			hex.EncodeToString(att.ExtraData), hex.EncodeToString(nonce))
	}
	lib.Print("Nonce from Quote matches expected nonce")

	sigL := tpm2.SignatureRSA{
		HashAlg:   tpm2.AlgSHA256,
		Signature: signature,
	}
	lib.Verbose("sigL: %v", sigL)

	// Read expected PCRs digest from disk
	pcrDigest := lib.Read(fmt.Sprintf("%s.bin", cicdDigestPath))

	//_, pcrHash, err := getPCRMap(tpm.HashAlgo_SHA256)
	//if err != nil {
	//	glog.Fatalf("Error getting PCRMap: %v", err)
	//}
	//glog.V(5).Infof("     sha256 of Expected PCR Value: --> %x", pcrHash)

	if !bytes.Equal(pcrDigest[:], att.AttestedQuoteInfo.PCRDigest) {
		lib.Fatal("Unexpected PCR hash Value Got 0x%s Expected: 0x%s",
			hex.EncodeToString(att.AttestedQuoteInfo.PCRDigest), hex.EncodeToString(pcrDigest[:]))
	}
	lib.Print("PCRs digest from Quote matches expected digest")

	// Verify AK signature
	// use the AK from the original attestation to verify the signature of the Attestation
	// rsaPub := rsa.PublicKey{E: int(tPub.RSAParameters.Exponent()), N: tPub.RSAParameters.Modulus()}
	akPublicKey := certs.ReadPublicKey(verifierAkPath)
	hsh := crypto.SHA256.New()
	hsh.Write(attestation)
	err = rsa.VerifyPKCS1v15(
		&akPublicKey,
		crypto.SHA256,
		hsh.Sum(nil),
		sigL.Signature,
	)
	if err != nil {
		lib.Fatal("rsa.VerifyPKCS1v15() failed: %v", err)
	}
	lib.Print("Quote signature is valid")
}

// === Verifier: verify quote2 =================================================

func VerifyQuote2(
	eventsLog []byte, // IN
	pcrs []int, // IN
	nonce []byte, // IN
	attestation []byte, // IN
	signature tpmutil.U16Bytes, // IN
	akPubPEM []byte, // In
) (
	isLegit bool,
	message string,
) {
	lib.Trace.Print("Even deeper.")
	parsedEventsLog, err := attest.ParseEventLog(eventsLog)
	if err != nil {
		lib.Fatal("attest.ParseEventLog() failed: %v", err)
	}

	// Compute expected PCR values
	allpcrs := [][32]byte{}
	for i := 0; i < 24; i++ {
		allpcrs = append(allpcrs, [32]byte{})
		lib.Verbose("PCR[%2d]: 0x%s", i, hex.EncodeToString(allpcrs[i][:]))
	}
	for _, e := range parsedEventsLog.Events(attest.HashAlg(tpm2.AlgSHA256)) {
		// sudo cat pcr.bin digest.bin | openssl dgst -sha256 -binary > futurepcr.bin
		i := e.Index
		allpcrs[i] = sha256.Sum256(append(allpcrs[i][:], e.Digest...))
		lib.Verbose("PCR[%2d]+0x%s => 0x%s", i,
			hex.EncodeToString(e.Digest), hex.EncodeToString(allpcrs[i][:]))
	}

	// Compute attestation digest
	lib.PRINT("=== INIT: PREDICT ATTESTATION DIGEST ===========================================")

	pcrsConcat := []byte{}
	for _, i := range pcrs {
		pcrsConcat = append(pcrsConcat, allpcrs[i][:]...)
	}
	pcrDigest := sha256.Sum256(pcrsConcat)

	lib.PRINT("=== VERIFIER: VERIFY QUOTE =====================================================")

	att, err := tpm2.DecodeAttestationData(attestation)
	if err != nil {
		lib.Fatal("DecodeAttestationData() failed: %v", err)
	}

	lib.Verbose("Attestation ExtraData (nonce): 0x%s ", hex.EncodeToString(att.ExtraData))
	lib.Verbose("Attestation PCR#: %v ", att.AttestedQuoteInfo.PCRSelection.PCRs)
	lib.Verbose("Attestation Hash: 0x%s ", hex.EncodeToString(att.AttestedQuoteInfo.PCRDigest))

	// Compare the nonce that is embedded within the attestation. This should
	// match the one we sent in earlier.
	if !bytes.Equal(nonce, att.ExtraData) {
		lib.Fatal("Nonce Value mismatch Got: (0x%s) Expected: (0x%s)",
			hex.EncodeToString(att.ExtraData), hex.EncodeToString(nonce))
	}
	lib.Print("Nonce from Quote matches expected nonce")

	sigL := tpm2.SignatureRSA{
		HashAlg:   tpm2.AlgSHA256,
		Signature: signature,
	}
	lib.Verbose("sigL: %v", sigL)

	if !bytes.Equal(pcrDigest[:], att.AttestedQuoteInfo.PCRDigest) {
		lib.Fatal("Unexpected PCR hash Value Got 0x%s Expected: 0x%s",
			hex.EncodeToString(att.AttestedQuoteInfo.PCRDigest), hex.EncodeToString(pcrDigest[:]))
	}
	lib.Print("PCRs digest from Quote matches expected digest")

	// Verify AK signature
	// use the AK from the original attestation to verify the signature of the Attestation
	// rsaPub := rsa.PublicKey{E: int(tPub.RSAParameters.Exponent()), N: tPub.RSAParameters.Modulus()}
	akPublicKey := certs.ReadPublicKeyPEM(akPubPEM)
	//fatal, msg, akPublicKey := certs.ReadPublicKeyPEM(akPubPEM)
	//if fatal {
	//	lib.Fatal("Cannot decode AK: %s", msg)
	//}
	hsh := crypto.SHA256.New()
	hsh.Write(attestation)
	err = rsa.VerifyPKCS1v15(
		&akPublicKey,
		crypto.SHA256,
		hsh.Sum(nil),
		sigL.Signature,
	)
	if err != nil {
		lib.Fatal("rsa.VerifyPKCS1v15() failed: %v", err)
	}
	lib.Print("Quote signature is valid")
	return true, "All good ;)"
}
