package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"rmazur.io/healthy"
	"time"
)

var configFile = flag.String("config-file", "", "configuration file path")

type httpCheckConfig struct {
	healthy.HttpCheck

	Period time.Duration `json:"period"`
	Flex time.Duration `json:"flex"`
}

type config struct {
	HttpChecks []*httpCheckConfig `json:"httpChecks"`
}

func main() {
	flag.Parse()
	var input *os.File
	closeInput := false
	if len(*configFile) == 0 {
		input = os.Stdin
		if stat, err := input.Stat(); err != nil || stat.Size() == 0 {
			log.Fatalf("Expected config file in stdin")
		}
	} else {
		var err error
		if input, err = os.Open(*configFile); err == nil {
			closeInput = true
		} else {
			log.Fatalf("Cannot open config file %s: %s\n", *configFile, err)
		}
	}

	var cfg config
	if err := json.NewDecoder(input).Decode(&cfg); err != nil {
		log.Fatalf("Cannot parse configuration: %s", err)
	} else if closeInput {
		_ = input.Close()
	}

	var checker healthy.Checker
	for _, hc := range cfg.HttpChecks {
		checker.AddTaskWithPeriod(&hc.HttpCheck, hc.Period, hc.Flex)
	}
	checker.Run(context.Background())

	stopSignal := make(chan os.Signal)
	signal.Notify(stopSignal, os.Interrupt)
	<-stopSignal
	checker.Stop()
}
