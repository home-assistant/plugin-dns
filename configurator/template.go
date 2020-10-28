package main

import (
	"bytes"
	"log"
	"strings"
	"text/template"
)

func renderCoreFile(config *CoreDNSConfig, coreFile string) []byte {
	buf := &bytes.Buffer{}

	// generate template
	coreTemplate := template.New("corefile").Funcs(template.FuncMap{"Join": strings.Join})
	template.Must(coreTemplate.ParseFiles(coreFile))

	// render
	err := coreTemplate.Execute(buf, *config)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}
