package main

import (
	"flag"
	"io/ioutil"
	"log"
)

func main() {
	configFile := flag.String("conf", "", "Config json file")
	coreFile := flag.String("template", "", "Corefile template file")
	outFile := flag.String("out", "", "Output file")

	// parse command lines
	flag.Parse()

	// Get config
	config := readConfigFile(*configFile)

	// Add fallback
	config.Locals = append(config.Locals, "dns://127.0.0.11")

	// create & write corefile
	data := renderCoreFile(config, *coreFile)
	err := ioutil.WriteFile(*outFile, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
