package util

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"text/template"
)

// GenerateUUID that will be used to generate uuid for variable email part (<CHANGE> part)
func GenerateUUID(length int) string {
	id := uuid.New()
	return id.String()[:length]
}

// WriteToFile populates csv file with matching users and ids
func WriteToFile(user, emails, url []string) {
	f, err := os.Create("targets_links.csv")
	if err != nil {
		log.Fatalf("Error creating file: %v\n", err)
	}
	defer f.Close()

	// Write header
	fmt.Fprintf(f, "%s,%s,%s\n", "Name", "Email", "Link sent")
	for i := range user {
		fmt.Fprintf(f, "%s,%s,%s\n", user[i], emails[i], url[i])
	}
}

// Result will hold data that will be written to final file that correlates uuid, username, password, name and email
type Result struct {
	UUID     string
	Username string
	Password string
	Email    string
	Name     string
	URL      string
}

const templ = `Usernames and passwords collected from Modlishka corelated with lateralus:
========================={{ range . }}
Name: {{ .Name }}
Email: {{ .Email }}
URL: {{ .URL }}
UUID: {{ .UUID }}
Username: {{ .Username}}
Password: {{ .Password }}
========================={{ end }}
`

// ParseModlishka parses control file from Modlishka and returns results
func ParseModlishka(controlDBFile string) {

	var results []Result

	data, err := ioutil.ReadFile(controlDBFile)
	if err != nil {
		log.Fatalf("Error reading control db file %s: %v\n", controlDBFile, err)
	}

	r := regexp.MustCompile(`{"UUID":"(?P<UUID>([^"]+))","Username":"(?P<USERNAME>[^"]+)","Password":"(?P<PASSWORD>[^"]+)"`)
	matches := r.FindAllStringSubmatch(string(data), -1)

	// Fill Result with UUID, Password and Username from control db file
	for _, v := range matches {
		if !uuidExists(results, v[1]) {
			results = append(results, Result{
				UUID:     v[1],
				Username: v[3],
				Password: v[4],
			})
		}
	}

	// Read targets_links.csv and fill the rest of the struct
	csvFile, err := os.Open("targets_links.csv")
	if err != nil {
		log.Fatalf("Error opening targets_links.csv: %v\n", err)
	}

	reader := csv.NewReader(csvFile)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			csvFile.Close()
			log.Fatalf("Error reading csv file: %v\n", err)
		}
		if uuid := extractUUID(record[2]); uuid != "" {
			for i, v := range results {
				if v.UUID == uuid {
					results[i].Name = record[0]
					results[i].Email = record[1]
					results[i].URL = record[2]
				}
			}
		}
	}
	csvFile.Close()

	t, _ := template.New("Results").Parse(templ)

	f, err := os.OpenFile("results.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening results.txt file: %v\n", err)
	}

	t.Execute(f, results)
}

func extractUUID(url string) string {
	// b629daf3-362f-497b-8321-f8e1b5c0526c
	r := regexp.MustCompile("[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}")
	uuid := r.FindString(url)
	return uuid
}

func uuidExists(res []Result, uuid string) bool {
	for _, a := range res {
		if a.UUID == uuid {
			return true
		}
	}
	return false
}
