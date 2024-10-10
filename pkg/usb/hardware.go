/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package usb

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.bug.st/serial"
)

type Hardware struct {
	port       serial.Port
	previousOk bool // did the previous command end in "OK"
	timeout    time.Duration
}

// Open the port based on the environment variable.
func (h *Hardware) OpenPortFromEnv(timeout time.Duration) error {
	portPath, err := h.EnvPort()
	if err != nil {
		return nil // don't set the port or error
	}
	h.previousOk = true // assume it was fine at the start
	return h.OpenPort(portPath, timeout)
}

func (h *Hardware) SetTimeout(timeout time.Duration) error {
	h.timeout = timeout
	if h.port == nil {
		return nil
	}
	return h.port.SetReadTimeout(timeout)
}

func (h *Hardware) PortSet() bool {
	return h.port != nil
}

// Open the port.
func (h *Hardware) OpenPort(path string, timeout time.Duration) error {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	var err error
	h.port, err = serial.Open(path, mode)
	if err != nil {
		return err
	}
	h.timeout = timeout
	return h.port.SetReadTimeout(timeout)
}

// Close the port, if open.
func (h *Hardware) Close() error {
	if h.port == nil {
		return nil
	}
	return h.port.Close()
}

func (h *Hardware) encode(bin []byte) string {
	return base64.StdEncoding.EncodeToString(bin)
}

func (h *Hardware) resetInput(t *testing.T) error {
	if h.previousOk {
		return nil
	}
	maxAttempts := 3
	h.previousOk = false
	var out []byte
	for i := 0; i < maxAttempts; i++ {
		t.Log("resetting usb input")
		h.port.SetDTR(false) // send reset signal to many dev boards
		h.port.SetDTR(true)
		time.Sleep(100 * time.Millisecond)   // wait for input
		h.port.ResetInputBuffer()            // reset the input buffer
		_, err := h.port.Write([]byte("\n")) // write a newline, esp32 should print " ok\r\n"
		if err != nil {
			return err
		}
		var previous []byte
		// read until we no longer have input
		for {
			buf := make([]byte, 65536)
			time.Sleep(10 * time.Millisecond)            // wait for read buffer to fill
			h.port.SetReadTimeout(10 * time.Millisecond) // temporarily decrease timeout
			n, _ := h.port.Read(buf)                     // read!
			h.port.SetReadTimeout(h.timeout)             // put the timeout back
			out = buf[:n]
			if n == 0 { // no bytes, we read it all
				if strings.HasSuffix(string(previous), " OK\r\n") || strings.HasSuffix(string(previous), " OK\n") {
					h.previousOk = true
					return nil
				}
				break
			}
			previous = make([]byte, n)
			copy(previous, out)
		}

	}
	return fmt.Errorf("exceeded max number of reset attempts, last message: %s %v", out, out)
}

// execute the binary while returning the unprocessed outputs
func (h *Hardware) executeRaw(payload string, t *testing.T) (string, error) {
	if h.port == nil {
		return "", fmt.Errorf("port not set up")
	}

	err := h.resetInput(t)
	if err != nil {
		return "", err
	}
	h.previousOk = false // default to not okay
	_, err = h.port.Write([]byte(payload + "\n"))
	if err != nil {
		return "", err
	}
	start := payload + " "

	str := ""
	buf := make([]byte, 1)
	for {
		read, err := h.port.Read(buf)
		if err != nil {
			return "", err
		}
		if read != 1 {
			return "", fmt.Errorf("read unexpected number of bytes: %d", read)
		}
		if buf[0] == '\r' {
			continue
		}
		if buf[0] == '\n' {
			if !strings.HasPrefix(str, start) {
				return "", fmt.Errorf("ulp app not sent correctly: %v", str)
			}
			str = str[len(start):]
			return str, nil
		}
		str += string(buf[0])
	}
}

// Execute a binary `bin`
func (h *Hardware) Execute(bin []byte, t *testing.T) (string, error) {
	encoded := h.encode(bin)
	maxAttempts := 5 // the maximum number of test attempts
	usbError := true
	var err error
	var words string
	for i := 0; i < maxAttempts; i++ {
		words, err = h.executeRaw(encoded, t)
		if err != nil { // an actual error occurred while executing. try again.
			h.previousOk = false
			continue
		}
		if strings.HasSuffix(words, " OK") {
			h.previousOk = true // the last was okay!
			words = words[:len(words)-3]
			usbError = false
			break // exit loop early
		}
		// some problem occurred while testing. try again.
		h.previousOk = false
		words = words[:len(words)-4]
	}
	if err != nil {
		return "", err
	}
	if usbError {
		return "", fmt.Errorf("error from test app: \"%v\"", strconv.Quote(words))
	}

	return words, nil
}

func (h *Hardware) EnvPort() (string, error) {
	port_env := "ESP_PORT"
	path := os.Getenv(port_env)
	if path == "" {
		return "", fmt.Errorf(port_env + " not set")
	}
	return path, nil
}
