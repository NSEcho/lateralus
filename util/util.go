package util

import (
	"github.com/google/uuid"
)

// GenerateUUID that will be used to generate uuid for variable email part (<CHANGE> part)
func GenerateUUID(length int) string {
	id := uuid.New()
	return id.String()[:length]
}
