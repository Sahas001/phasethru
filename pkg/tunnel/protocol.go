package tunnel

import (
	"encoding/json"
	"io"
)

// HandshakeRequest is sent by the client CLI to the server daemon when initiating the control connection.
type HandshakeRequest struct {
	RequestedSubdomain string `json:"requested_subdomain"`
}

// HandshakeResponse is sent by the server daemon to the client CLI in response to a connection attempt.
type HandshakeResponse struct {
	Subdomain string `json:"subdomain"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

// WriteHandshake serializes a value to JSON, appends a newline, and writes it to the writer.
func WriteHandshake(w io.Writer, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = w.Write(append(data, '\n'))
	return err
}

// ReadLine reads bytes one-by-one from the reader until a newline character is encountered.
// This is critical to prevent standard buffered readers from consuming extra bytes meant
// for the subsequent Yamux session payload.
func ReadLine(r io.Reader) (string, error) {
	var buf []byte
	oneByte := make([]byte, 1)
	for {
		n, err := r.Read(oneByte)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}
		if oneByte[0] == '\n' {
			break
		}
		buf = append(buf, oneByte[0])
	}
	return string(buf), nil
}

// ReadHandshake reads a newline-terminated JSON line from the reader and parses it into the target struct.
func ReadHandshake(r io.Reader, val interface{}) error {
	line, err := ReadLine(r)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(line), val)
}
