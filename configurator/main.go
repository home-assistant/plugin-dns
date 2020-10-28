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

	config := readConfigFile(*configFile)
	data := renderCoreFile(config, *coreFile)

	err := ioutil.WriteFile(*outFile, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
