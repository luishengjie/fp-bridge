package event

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewID() (string, error) {
	var bytes [16]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate random event ID: %w", err)
	}

	return "evt_" + hex.EncodeToString(bytes[:]), nil
}
