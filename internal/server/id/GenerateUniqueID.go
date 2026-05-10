package id

import "github.com/google/uuid"

func GenerateUniqueID() (string, error) {
	id := uuid.New()
	return id.String(), nil
}
