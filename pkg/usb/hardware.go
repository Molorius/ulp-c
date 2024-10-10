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
	"time"

	"go.bug.st/serial"
)

type Hardware struct {
	port serial.Port
}

// Open the port based on the environment variable.
func (h *Hardware) OpenPortFromEnv() error {
	portPath, err := h.EnvPort()
	if err != nil {
		return nil // don't set the port or error
	}
	return h.OpenPort(portPath)
}

func (h *Hardware) PortSet() bool {
	return h.port != nil
}

// Open the port.
func (h *Hardware) OpenPort(path string) error {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	var err error
	h.port, err = serial.Open(path, mode)
	return err
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

// execute the binary while returning the unprocessed outputs
func (h *Hardware) executeRaw(payload string) (string, error) {
	if h.port == nil {
		return "", fmt.Errorf("port not set up")
	}
	var err error

	h.port.ResetInputBuffer()
	h.port.ResetOutputBuffer()
	h.port.SetReadTimeout(2 * time.Second)

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
func (h *Hardware) Execute(bin []byte) (string, error) {
	encoded := h.encode(bin)
	words, err := h.executeRaw(encoded)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(words, " OK") {
		words = words[:len(words)-3]
		return words, nil
	}
	if strings.HasSuffix(words, " ERR") {
		words = words[:len(words)-4]
		return words, fmt.Errorf("error from ulp app: \"%v\"", words)
	}
	return "", fmt.Errorf("error from test app: \"%v\"", strconv.Quote(words))
}

func (h *Hardware) EnvPort() (string, error) {
	port_env := "ESP_PORT"
	path := os.Getenv(port_env)
	if path == "" {
		return "", fmt.Errorf(port_env + " not set")
	}
	return path, nil
}
