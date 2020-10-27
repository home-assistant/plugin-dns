package main

import "flag"

func main() {
	configFile := flag.String("conf", "", "Config json file")
	coreFile := flag.String("template", "", "Corefile template file")
	outFile := flag.String("out", "", "Output file")

	config := readConfigFile(configFile)
	data := renderCoreFile(config, coreFile)

}
