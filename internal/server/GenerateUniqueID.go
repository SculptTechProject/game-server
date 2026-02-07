package server

import "github.com/google/uuid"

func GenerateUniqeID() (string, error) {
	id := uuid.New()
	return id.String(), nil
}
