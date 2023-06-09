// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"fmt"

	"github.com/google/go-tpm/tpmutil"

	"main/src/lib"
)

// === Verifier: verify TPM quote ==============================================

func ExtVerifyTpmQuote(
	cicdPredictionPath string, // IN
	pcrs []int, // IN
	nonce []byte, // IN
	attestation []byte, // IN
	signature tpmutil.U16Bytes, // IN
	akPub string, // IN
) (
	isLegit bool,
	message string,
) {
	// Local panic handler
	defer func() {
		if e := recover(); e != nil {
			// In the recovery process, we override the value of ExtVerifyTpmQuote named return parameters
			// as appropriate to convey the nature of the problem back to the browser window.
			isLegit = false
			switch x := e.(type) {
			case string:
				message = x
			default:
				message = "unknown error"
			}
		}
	}()

	// Retrieve events log
	eventsLog := lib.Read(fmt.Sprintf("%s.bin", cicdPredictionPath))
	lib.Trace.Print("In deeper.")
	return VerifyQuote2(
		eventsLog,     // IN
		pcrs,          // IN
		nonce,         // IN
		attestation,   // IN
		signature,     // IN
		[]byte(akPub), // IN
	)
}
