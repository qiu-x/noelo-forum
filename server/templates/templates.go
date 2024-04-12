package templates

import (
	"log"
	"os"
	"text/template"
)

func CreateTemplate(tempPath string) *template.Template {
	tempRaw, err := os.ReadFile(tempPath)
	if err != nil {
		log.Fatal(err)
	}
	tempStr := string(tempRaw)
	temp, err := template.New("temp").Parse(tempStr)
	if err != nil {
		log.Fatal(err)
	}
	return temp
}
