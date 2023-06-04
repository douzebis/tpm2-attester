// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"

	"main/src/certs"
	"main/src/lib"
	"main/src/steps"
	"main/src/teepeem"
)

var (
	tpmPath = flag.String("tpm-path", "/dev/tpmrm0", "Path to the TPM device (character device or a Unix socket).")
	flush   = flag.String("flush", "all", "Flush contexts, must be oneof transient|saved|loaded|all")
)

func main() {
	flag.Parse()

	lib.PRINT("### INIT: CREATE CA ROOT FOR MANUFACTURER AND OWNER ############################")

	// Create certificate for Manufacturer CA
	lib.PRINT("=== MANUFACTURER: CREATE MANUFACTURER CA CERT ==================================")
	certs.CreateCACert(
		"Manufacturer",
		"Manufacturer/manufacturer-ca",
	)

	// Create certificate for Owner CA
	lib.PRINT("=== OWNER: CREATE OWNER CA CERT ================================================")
	certs.CreateCACert(
		"Owner",
		"Owner/owner-ca",
	)

	lib.PRINT("### CICD: PREDICT DIGESTS ######################################################")
	// In this mock-up we cheat by reading the digests from the events log.
	// Normally the CICD should predict the digests from the assets it builds.

	// Retrieve and save TPM events log
	lib.PRINT("=== INIT: RETRIEVE EVENTS LOG ==================================================")
	//eventsLog, err := client.GetEventLog(rwc)
	//if err != nil {
	//	lib.Fatal("client.GetEventLog(): %v", err)
	//}
	eventsLog := lib.Read("/sys/kernel/security/tpm0/binary_bios_measurements")
	lib.Write("CICD/cicd-prediction.bin", eventsLog, 0644)

	lib.PRINT("### MANUFACTURER: CREATE TPM CERT ##############################################")

	// Open TPM
	lib.PRINT("=== INIT: OPEN TPM =============================================================")
	rwc := teepeem.OpenFlush(*tpmPath, *flush)
	defer rwc.Close()

	// Read and save TPM EK Pub
	lib.PRINT("=== INIT: RETRIEVE EK PUB ======================================================")
	steps.GetEKPub(
		rwc,
		"Manufacturer/ek", // OUT
	)

	// Create TPM EK Cert
	lib.PRINT("=== INIT: CREATE EK CERT =======================================================")
	certs.CreateEKCert(
		"Manufacturer/ek", // IN
		"id: Google",
		"Shielded VM vTPM",
		"id: 00010001",
		"Manufacturer/manufacturer-ca", // IN
		"Manufacturer/ek",              // OUT
	)

	// Attestor: retrieve EK Pub from TPM
	steps.GetEKPub(
		rwc,
		"Attestor/ek", // OUT
	)

	// Verifier: verify EK Pub with Manufacturer EK Cert
	steps.VerifyEKPub(
		"Attestor/ek",                  // IN
		"Manufacturer/ek",              // IN
		"Manufacturer/manufacturer-ca", // IN
		"Verifier/ek",                  // OUT
	)

	// Verifier/Owner: create Owner EK Cert
	certs.CreateEKCert(
		"Verifier/ek",      // IN
		"id: Google",       // IN
		"Shielded VM vTPM", // IN
		"id: 00010001",     // IN
		"Owner/owner-ca",   // IN
		"Verifier/ek",      // OUT
	)

	// Attestor: create AK
	steps.CreateAK(
		rwc,
		"Attestor/ek", // IN
		"Attestor/ak", // OUT
	)

	// Verifier: generate credential challenge
	steps.GenerateCredential(
		"Attestor/ak",         // IN
		"Verifier/ek",         // IN
		"Verifier/nonce",      // OUT
		"Verifier/credential", // OUT
	)

	// Attestor: activate credential
	steps.ActivateCredential(
		rwc,                   // IN
		"Verifier/credential", // IN
		"Attestor/ek",         // IN
		"Attestor/ak",         // IN
		"Attestor/attempt",    // OUT
	)

	// Verifier: verify credential
	steps.VerifyCredential(
		"Attestor/attempt", // IN
		"Verifier/nonce",   // IN
		"Attestor/ak",      // IN
		"Verifier/ak",      // OUT
	)

	// Verifier/Owner: create Owner AK Cert
	certs.CreateAKCert(
		"Verifier/ak",    // IN
		"TPM AK",         // IN
		"Owner/owner-ca", // IN
		"Verifier/ak",    // OUT
	)

}
