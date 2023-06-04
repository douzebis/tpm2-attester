// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"

	"github.com/golang/glog"
)

// constants for Logger
var (
	// Whether lib traces & errors go through "log" or "glog"
	UseLog bool = false
	// Trace logs general information messages.
	Trace *log.Logger
	// Error logs error messages.
	Error *log.Logger
)

const (
	// See https://pkg.go.dev/github.com/ccpaging/nxlog4go@v2.0.3+incompatible/console#section-readme
	RESET      = "\033[0m" // No Color
	BOLD_WHITE = "\033[1;37m"
	BLUE       = "\033[0;34m"
	GREEN      = "\033[0;32m"
	ORANGE     = "\033[0;33m"
	PURPLE     = "\033[0;35m"
	RED        = "\033[0;31m"
)

func Fatal(format string, params ...interface{}) {
	message := fmt.Sprintf(format, params...)
	panic(message)
}

func PRINT(format string, params ...interface{}) {
	message := fmt.Sprintf(format, params...)
	if UseLog {
		Trace.Printf("%s", message)
	} else {
		glog.V(0).Infof("%s%s%s", BOLD_WHITE, message, RESET)
	}
}

func Print(format string, params ...interface{}) {
	message := fmt.Sprintf(format, params...)
	if UseLog {
		Trace.Printf("%s", message)
	} else {
		glog.V(0).Infof("%s%s%s", ORANGE, message, RESET)
	}
}

func Comment(format string, params ...interface{}) {
	message := fmt.Sprintf(format, params...)
	if UseLog {
		Trace.Printf("%s", message)
	} else {
		glog.V(5).Infof("%s%s%s", GREEN, message, RESET)
	}
}

func Verbose(format string, params ...interface{}) {
	message := fmt.Sprintf(format, params...)
	if UseLog {
		Trace.Printf("%s", message)
	} else {
		glog.V(10).Infof("%s%s%s", BLUE, message, RESET)
	}
}

func Read(
	path string,
) (
	data []byte,
) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		Fatal("ioutil.ReadFile() failed: %v", err)
	}

	Comment("Read %s", path)

	return data
}

func Write(
	path string,
	data []byte,
	perm fs.FileMode,
) {
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		Fatal("ioutil.WriteFile() failed: %v", err)
	}
	Comment("Wrote %s", path)
}
