// SPDX-License-Identifier: Apache-2.0

/*
* Original author of the native messaging "harness"
* @Author: J. Farley
* @Date: 2019-05-19
* @Description: Basic chrome native messaging host example.
 */
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"runtime/debug"
	"unsafe"

	"main/src/lib"
	"main/src/steps"
	"main/src/teepeem"
)

var (
	// devicePath is needed because firefox behaves differently than chromium
	// - in chromium, the host process is started with the current directory
	//   set to the directory that contains the host binary
	// - in firefox, the host process is started with the current directory
	//   set to the home directory of the process owner
	devicePath = "ATTESTER_DEVICE_PATH/"  // use absolute path so that both chromium and firefox will work
	tpmPath = flag.String("tpm-path", "/dev/tpmrm0", "Path to the TPM device (character device or a Unix socket).")
	flush   = flag.String("flush", "all", "Flush contexts, must be oneof transient|saved|loaded|all")
	rwc     io.ReadWriteCloser
)

//// constants for Logger
//var (
//	// Trace logs general information messages.
//	Trace *log.Logger
//	// Error logs error messages.
//	Error *log.Logger
//)

// nativeEndian used to detect native byte order
var nativeEndian binary.ByteOrder

// bufferSize used to set size of IO buffer - adjust to accommodate message payloads
var bufferSize = 8192

// IncomingMessage represents a message sent to the native host.
type IncomingMessage struct {
	Query       string    `json:"query"`
	Nonce       [32]byte  `json:"nonce"`
	Attestation [145]byte `json:"attestation"`
	Signature   [256]byte `json:"signature"`
	AkPub       string    `json:"ak-pub"`
	Pcrs        []int     `json:"pcrs"`
}

// OutgoingMessage respresents a response to an incoming message query.
type OutgoingMessage struct {
	Query       string    `json:"query"`
	Nonce       [32]byte  `json:"nonce"`
	Attestation [145]byte `json:"attestation"`
	Signature   [256]byte `json:"signature"`
	AkPub       string    `json:"ak-pub"`
	IsLegit     bool      `json:"is-legit"`
	Message     string    `json:"message"`
}

// Init initializes logger and determines native byte order.
func Init(traceHandle io.Writer, errorHandle io.Writer) {
	lib.Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	lib.Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// determine native byte order so that we can read message size correctly
	var one int16 = 1
	b := (*byte)(unsafe.Pointer(&one))
	if *b == 0 {
		nativeEndian = binary.BigEndian
	} else {
		nativeEndian = binary.LittleEndian
	}
}

func main() {
	// force all output through "log" (including "main/src/lib"'s)
	lib.UseLog = true

	// Global panic handler
	defer func() {
		if message := recover(); message != nil {
			lib.Print("%s", message)
			lib.Print("%s", debug.Stack())
		}
	}()

	file, err := os.OpenFile("/tmp/attester.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Init(os.Stdout, os.Stderr)
		lib.Error.Printf("Unable to create and/or open log file. Will log to Stdout and Stderr. Error: %v", err)
	} else {
		Init(file, file)
		// ensure we close the log file when we're done
		defer file.Close()
	}

	// Open TPM and Flush handles
	rwc = teepeem.OpenFlush(*tpmPath, *flush)
	defer rwc.Close()

	lib.Trace.Printf("Chrome native messaging host started. Native byte order: %v.", nativeEndian)
	read()
	lib.Trace.Print("Chrome native messaging host exited.")
}

// read Creates a new buffered I/O reader and reads messages from Stdin.
func read() {
	v := bufio.NewReader(os.Stdin)
	// adjust buffer size to accommodate your json payload size limits; default is 4096
	s := bufio.NewReaderSize(v, bufferSize)
	lib.Trace.Printf("IO buffer reader created with buffer size of %v.", s.Size())

	lengthBytes := make([]byte, 4)
	lengthNum := int(0)

	// we're going to indefinitely read the first 4 bytes in buffer, which gives us the message length.
	// if stdIn is closed we'll exit the loop and shut down host
	for b, err := s.Read(lengthBytes); b > 0 && err == nil; b, err = s.Read(lengthBytes) {
		// convert message length bytes to integer value
		lengthNum = readMessageLength(lengthBytes)
		lib.Trace.Printf("Message size in bytes: %v", lengthNum)

		// If message length exceeds size of buffer, the message will be truncated.
		// This will likely cause an error when we attempt to unmarshal message to JSON.
		if lengthNum > bufferSize {
			lib.Error.Printf("Message size of %d exceeds buffer size of %d. Message will be truncated and is unlikely to unmarshal to JSON.", lengthNum, bufferSize)
		}

		// read the content of the message from buffer
		content := make([]byte, lengthNum)
		_, err := s.Read(content)
		if err != nil && err != io.EOF {
			lib.Error.Fatal(err)
		}

		// message has been read, now parse and process
		parseMessage(content)
	}

	lib.Trace.Print("Stdin closed.")
}

// readMessageLength reads and returns the message length value in native byte order.
func readMessageLength(msg []byte) int {
	var length uint32
	buf := bytes.NewBuffer(msg)
	err := binary.Read(buf, nativeEndian, &length)
	if err != nil {
		lib.Error.Printf("Unable to read bytes representing message length: %v", err)
	}
	return int(length)
}

// parseMessage parses incoming message
func parseMessage(msg []byte) {
	iMsg := decodeMessage(msg)
	lib.Trace.Printf("Message received: %s", msg)
	path, err := os.Getwd()
	if err != nil {
		lib.Print("%v", err)
	}
	lib.Print("%s", path)

	// start building outgoing json message
	oMsg := OutgoingMessage{
		Query: iMsg.Query,
	}

	switch iMsg.Query {
	case "get-ak-pub":
		oMsg.AkPub = string(steps.ExtGetAkPub(devicePath+"Verifier/ak"))
	case "get-tpm-quote":
		nonce, attestation, signature := steps.ExtGetTpmQuote(
			rwc,
			devicePath,
			iMsg.Pcrs,
		)
		var byteArray [32]byte
		copy(byteArray[:], nonce)
		oMsg.Nonce = byteArray
		copy(oMsg.Attestation[:], attestation)
		copy(oMsg.Signature[:], signature)
	case "verify-tpm-quote":
		oMsg.IsLegit, oMsg.Message = steps.ExtVerifyTpmQuote(
			devicePath+"CICD/cicd-prediction", // IN
			iMsg.Pcrs,              // IN
			iMsg.Nonce[:],          // IN
			iMsg.Attestation[:],    // IN
			iMsg.Signature[:],      // IN
			iMsg.AkPub,             // IN
		)
	}
	send(oMsg)
}

// decodeMessage unmarshals incoming json request and returns query value.
func decodeMessage(msg []byte) IncomingMessage {
	var iMsg IncomingMessage
	err := json.Unmarshal(msg, &iMsg)
	if err != nil {
		lib.Error.Printf("Unable to unmarshal json to struct: %v", err)
	}
	return iMsg
}

// send sends an OutgoingMessage to os.Stdout.
func send(msg OutgoingMessage) {
	byteMsg := dataToBytes(msg)
	writeMessageLength(byteMsg)

	var msgBuf bytes.Buffer
	_, err := msgBuf.Write(byteMsg)
	if err != nil {
		lib.Error.Printf("Unable to write message length to message buffer: %v", err)
	}

	_, err = msgBuf.WriteTo(os.Stdout)
	if err != nil {
		lib.Error.Printf("Unable to write message buffer to Stdout: %v", err)
	}
}

// dataToBytes marshals OutgoingMessage struct to slice of bytes
func dataToBytes(msg OutgoingMessage) []byte {
	byteMsg, err := json.Marshal(msg)
	if err != nil {
		lib.Error.Printf("Unable to marshal OutgoingMessage struct to slice of bytes: %v", err)
	}
	return byteMsg
}

// writeMessageLength determines length of message and writes it to os.Stdout.
func writeMessageLength(msg []byte) {
	err := binary.Write(os.Stdout, nativeEndian, uint32(len(msg)))
	if err != nil {
		lib.Error.Printf("Unable to write message length to Stdout: %v", err)
	}
}
