package main

import (
	"bytes"
	"log"
	"template"
)

func renderCoreFile(config *CoreDNSConfig, coreFile string) []byte {
	buf := &bytes.Buffer{}
	coreTemplate := template.Must(template.ParseFile(coreFile))

	err := coreTemplate.Execute(buf, config)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}
