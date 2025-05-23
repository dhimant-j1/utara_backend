package utils

import (
	"bytes"
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a base64 encoded QR code image from the given data
func GenerateQRCode(data string) (string, error) {
	var buf bytes.Buffer

	// Generate QR code with medium recovery level and size 256
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", err
	}

	// Get PNG bytes
	png, err := qr.PNG(256)
	if err != nil {
		return "", err
	}

	// Encode to base64
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	_, err = encoder.Write(png)
	if err != nil {
		return "", err
	}
	encoder.Close()

	return "data:image/png;base64," + buf.String(), nil
}
