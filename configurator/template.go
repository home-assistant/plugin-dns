package main

import (
	"bytes"
	"log"
	"strings"
	"template"
)

func renderCoreFile(config *CoreDNSConfig, coreFile string) []byte {
	buf := &bytes.Buffer{}
	coreTemplate := template.Must(template.ParseFile(coreFile))

	// Add functions
	coreTemplate.Funcs(template.FuncMap{"Join": strings.Join})

	err := coreTemplate.Execute(buf, config)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}
