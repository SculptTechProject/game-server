package id

import (
	"crypto/rand"
	"math/big"
	"strconv"

	"github.com/google/uuid"
)

func GenerateUniqueID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func GenerateRoomCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return "1234"
	}
	return strconv.FormatInt(1000+n.Int64(), 10)
}
