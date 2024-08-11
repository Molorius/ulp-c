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

func (h *Hardware) encode(bin []byte) string {
	return base64.StdEncoding.EncodeToString(bin)
}

// execute the binary while returning the unprocessed outputs
func (h *Hardware) executeRaw(portPath string, payload string) (string, error) {
	var err error

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	h.port, err = serial.Open(portPath, mode)
	if err != nil {
		return "", fmt.Errorf("serial.Open fail with %s: %s", portPath, err)
	}
	defer h.port.Close()
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

// Execute a binary `bin` on path `portPath`
func (h *Hardware) Execute(portPath string, bin []byte) (string, error) {

	encoded := h.encode(bin)
	words, err := h.executeRaw(portPath, encoded)
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

// Execute a binary, obtain the path automatically.
func (h *Hardware) ExecuteEnv(bin []byte) (string, error) {
	port, err := h.EnvPort()
	if err != nil {
		return "", err
	}
	return h.Execute(port, bin)
}

func (h *Hardware) EnvPort() (string, error) {
	port_env := "ESP_PORT"
	path := os.Getenv(port_env)
	if path == "" {
		return "", fmt.Errorf(port_env + " not set")
	}
	return path, nil
}
