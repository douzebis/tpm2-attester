// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"

	"main/src/lib"
	"main/src/teepeem"
)

// === Attestor: perform quote =================================================

func PerformQuote(
	rw io.ReadWriter,
	attestorEkPath string, // IN
	attestorAkPath string, // IN
	pcrs []int, // IN
	verifierNoncePath string, // IN
	attestorQuotePath string, // OUT
) (
	attestation []byte,
	signature tpmutil.U16Bytes,
) {

	lib.PRINT("=== ATTESTOR: PERFORM QUOTE ====================================================")
	lib.Trace.Print("0")
	lib.Trace.Printf("attestorEkPath %s", attestorEkPath)

	// Load EK
	ek := teepeem.LoadEK(
		rw,
		attestorEkPath,
	)
	lib.Trace.Print("1")
	defer tpm2.FlushContext(rw, ek)
	lib.Trace.Print("2")

	// Load AK
	ak, _ := teepeem.LoadAK(
		rw,
		ek,
		attestorAkPath, // IN
	)
	lib.Trace.Print("3")
	defer tpm2.FlushContext(rw, ak)
	lib.Trace.Print("4")

	// Load nonce
	nonce := lib.Read(fmt.Sprintf("%s.bin", verifierNoncePath))
	lib.Trace.Print("5")

	// Perform quote
	pcrSelection := tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: pcrs,
	}
	lib.Trace.Print("6")
	attestation, sig, err := tpm2.Quote(
		rw,
		ak,
		"", // emptyPassword
		"", // emptyPassword
		nonce,
		pcrSelection,
		tpm2.AlgNull,
	)
	lib.Trace.Print("7")
	if err != nil {
		lib.Fatal("tpm2.Quote() failed: %v", err)
	}
	lib.Trace.Print("8")
	signature = sig.RSA.Signature
	lib.Verbose("     Quote Hex %v", hex.EncodeToString(attestation))
	lib.Verbose("     Quote Sig %v", hex.EncodeToString(signature))

	lib.Trace.Print("9")
	// Write quote to disk
	lib.Write(fmt.Sprintf("%s-attest.bin", attestorQuotePath), attestation, 0644)
	lib.Write(fmt.Sprintf("%s-signature.bin", attestorQuotePath), signature, 0644)

	return attestation, signature
}
