// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"fmt"

	"main/src/lib"
)

// === Verifier: get AK pub ====================================================

func ExtGetAkPub(
	akPubPath string, // IN
) (
	akPubPEM []byte,
) {
	// Read AK pub from disk
	publicKeyPEM := lib.Read(fmt.Sprintf("%s.pub", akPubPath))
	lib.Verbose("publicKeyPEM: %v", publicKeyPEM)

	return publicKeyPEM
}
