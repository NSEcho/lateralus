package models

type OutputFormat string

const (
	OutputTypeJSON OutputFormat = "json"
	OutputTypeXML  OutputFormat = "xml"
)

type Target struct {
	Name  string
	Email string
}

type SendingData struct {
	Target
	Body         string
	AttackerName string
	URL          string
	Custom       string
}

type Result struct {
	StartTime    string
	EndTime      string
	Subject      string
	From         string
	AttackerName string
	URL          string
	Custom       string
	Targets      []SendingData
}
