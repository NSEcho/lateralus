package util

import "github.com/google/uuid"

// GenerateUUID will be used to generate random part of url
func GenerateUUID(length int) string {
	id := uuid.New()
	return id.String()[:length]
}
