package proclimit

import (
	"errors"
	"github.com/google/uuid"
)

// randomName generates a random UUID (v4)
func randomName() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("failed to generate uuid")
	}
	return u.String(), nil
}
