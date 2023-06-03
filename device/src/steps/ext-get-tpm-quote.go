// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"io"

	"main/src/lib"
)

// === Attestor: get TPM quote =================================================

func ExtGetTpmQuote(
	rwc io.ReadWriteCloser, // IN
	pcrs []int, // IN
) (
	nonce []byte,
	attestation []byte,
	signature []byte,
) {

	// Verifier: request PCR quote
	nonce = RequestQuote(
		"Verifier/nonce-quote", // OUT
	)
	lib.Trace.Print("AA")
	// Attestor: perform PCR quote
	attestation, signature = PerformQuote(
		rwc,
		"Attestor/ek",          // IN
		"Attestor/ak",          // IN
		pcrs,                   // IN
		"Verifier/nonce-quote", // IN
		"Attestor/quote",       // OUT
	)
	lib.Trace.Print("BB")

	return nonce, attestation, signature
}
