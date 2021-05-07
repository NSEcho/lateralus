package util

import "github.com/google/uuid"

func GenerateUUID(length int) string {
	id := uuid.New()
	return id.String()[:length]
}
