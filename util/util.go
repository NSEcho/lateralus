package util

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"os"
)

// GenerateUUID that will be used to generate uuid for variable email part (<CHANGE> part)
func GenerateUUID(length int) string {
	id := uuid.New()
	return id.String()[:length]
}

// WriteToFile populates csv file with matching users and ids
func WriteToFile(user, url []string) {
	f, err := os.Create("targets_links.csv")
	if err != nil {
		log.Fatalf("Error creating file: %v\n", err)
	}
	defer f.Close()
	for i := range user {
		fmt.Fprintf(f, "%s,%s\n", user[i], url[i])
	}
}
