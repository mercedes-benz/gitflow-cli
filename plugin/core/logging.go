/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Logging bit flags for controlling logging behavior for all repository operations.
const (
	_ Logging = 1 << iota
	Off
	StdErr
	StdOut
	CmdLine
	Output
)

type (
	// Logging controls logging behavior for all repository operations.
	Logging int
)

// LoggingSetting controls logging behavior for all repository operations.
const loggingSetting = "logging"

// LoggingNames maps logging flags to their names.
var loggingNames = map[Logging]string{
	Off:     "off",
	StdErr:  "stderr",
	StdOut:  "stdout",
	CmdLine: "cmdline",
	Output:  "output",
}

// Internal flags for controlling core package behavior.
var loggingFlags Logging = StdOut | CmdLine | Output

// Log a message to Go standard logging based on logging flags and variadic arguments.
func Log(message ...any) {
	println := func() {
		for _, msg := range message {
			switch msg := msg.(type) {
			case string:
				if len(msg) > 0 && (loggingFlags&CmdLine != 0 || loggingFlags&Output != 0) {
					log.Println(msg)
				}

			case *exec.Cmd:
				if msg != nil && len(msg.String()) > 0 && loggingFlags&CmdLine != 0 {
					log.Println(msg.String())
				}

			case []byte:
				if len(msg) > 0 && loggingFlags&Output != 0 {
					output := strings.TrimRight(string(msg), "\n\r")
					log.Println(output)
				}

			case error:
				if msg != nil && len(msg.Error()) > 0 && loggingFlags&Output != 0 {
					log.Println(msg.Error())
				}

			default:
				if msg != nil && len(fmt.Sprintf("%v", msg)) > 0 && loggingFlags&Output != 0 {
					log.Println(msg)
				}
			}
		}
	}

	if loggingFlags&StdErr != 0 {
		log.SetOutput(os.Stderr)
		println()
	}

	if loggingFlags&StdOut != 0 {
		log.SetOutput(os.Stdout)
		println()
	}
}

// String representation of a logging flag (only one allowed at a time).
func (l Logging) String() string {
	return loggingNames[l]
}
