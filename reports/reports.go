package reports

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/lateralusd/lateralus/models"
)

const tpl = `Start time:     {{ .StartTime }}
End time:       {{ .EndTime }}

Mail data:
========================================
Mail Subject: 	{{ .Subject }}
From field: 	{{ .From }}
AttackerName: 	{{ .AttackerName }}
URL: 		{{ .URL }}
Custom: 	{{ .Custom }}

Targets:
========================================
Total: 			{{ len .Targets }}
Table in format NAME, EMAIL, URL
----------------------------------------{{ range .Targets }}
{{ .Name | printf "%-20s"}} | {{ .Email | printf "%-50s"}} | {{ .URL }}
{{end}}`

// CreateReport will create report in json, xml or good ol text/template format
func CreateReport(filename, templatePath string, format models.OutputFormat, res *models.Result) error {
	var err error
	switch format {
	case models.OutputTypeJSON:
		err = createOutput(filename, models.OutputTypeJSON, res)
	case "xml":
		err = createOutput(filename, models.OutputTypeXML, res)
	default:
		err = createTemplate(filename, templatePath, res)
	}

	if err != nil {
		return fmt.Errorf("CreateReport: %v", err)
	}

	return nil
}

func createOutput(filename string, outputType models.OutputFormat, res *models.Result) error {
	var data []byte
	var err error

	if outputType == models.OutputTypeJSON {
		data, err = json.MarshalIndent(res, "", " ")
	} else {
		data, err = xml.MarshalIndent(res, "", " ")
	}

	if err != nil {
		return fmt.Errorf("createOutput: %v", err)
	}

	file := filename + ".rep"

	if err := ioutil.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("createOutput: %v", err)
	}

	return nil
}

func createTemplate(filename, templatePath string, res *models.Result) error {
	var t *template.Template
	var err error

	if templatePath == "" {
		t, err = template.New("").Parse(tpl)
		if err != nil {
			return fmt.Errorf("createTemplate: %v", err)
		}
	} else {
		t, err = template.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("createTemplate: %v", err)
		}
	}

	file := filename + ".rep"

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("createTemplate: %v", err)
	}

	if err := t.Execute(f, res); err != nil {
		return fmt.Errorf("createTemplate: %v", err)
	}

	return nil
}
