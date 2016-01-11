package main

import (
	"calculator/reaper"
	"flag"
	"os"
	"utils/logger"
)

func usage() {
	logger.Println("Usage: calculator [-c config_file]")
	os.Exit(0)
}

func main() {

	var config = flag.String("c", "", "domain tag config file")
	flag.Usage = usage
	flag.Parse()
	logger.Info("start to run raper service.")
	reaper.StartService(*config)
}
