package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"rmazur.io/healthy"
	"strings"
	"time"
)

var configFile = flag.String("config-file", "", "configuration file path")

type duration time.Duration

func (d *duration) UnmarshalJSON(data []byte) error {
	res, err := time.ParseDuration(strings.Trim(string(data), "\""))
	*d = duration(res)
	return err
}

type httpCheckConfig struct {
	Url                string `json:"url"`
	ExpectedStatusCode int    `json:"expectedStatusCode"`

	Timeout duration `json:"timeout"`

	Period duration `json:"period"`
	Flex   duration `json:"flex"`
}

type config struct {
	HttpChecks          []*httpCheckConfig `json:"httpChecks"`
	ReportFailuresCount int                `json:"reportFailuresCount"`
	FirstRetryDelay     duration           `json:"firstRetryDelay"`
	Twillio             struct {
		AccountId string `json:"accountId"`
		AuthToken string `json:"authToken"`
		From      string `json:"from"`
		To        string `json:"to"`
	} `json:"twillio"`
}

type logNotifier struct {
	*log.Logger
}

func (l *logNotifier) Notify(taskName string, e error) {
	l.Printf("New failure detected for task %s - %s", taskName, e)
}

type compositeNotifier []healthy.Notifier

func (c compositeNotifier) Notify(taskName string, e error) {
	for _, n := range c {
		n.Notify(taskName, e)
	}
}

func main() {
	lg := log.New(os.Stdout, "[healthy] ", log.LstdFlags)

	flag.Parse()
	var input *os.File
	closeInput := false
	if len(*configFile) == 0 {
		input = os.Stdin
		if stat, err := input.Stat(); err != nil || stat.Size() == 0 {
			lg.Fatalf("Expected config file in stdin")
		}
	} else {
		var err error
		if input, err = os.Open(*configFile); err == nil {
			closeInput = true
		} else {
			lg.Fatalf("Cannot open config file %s: %s\n", *configFile, err)
		}
	}

	var cfg config
	if err := json.NewDecoder(input).Decode(&cfg); err != nil {
		lg.Fatalf("Cannot parse configuration: %s", err)
	} else if closeInput {
		_ = input.Close()
	}

	notifier := compositeNotifier{&logNotifier{lg}}
	if len(cfg.Twillio.AccountId) > 0 && len(cfg.Twillio.AuthToken) > 0 &&
		len(cfg.Twillio.From) > 0 && len(cfg.Twillio.To) > 0 {
		notifier = append(notifier, &twillio{
			accountSid:     cfg.Twillio.AccountId,
			authToken:      cfg.Twillio.AuthToken,
			senderNumber:   cfg.Twillio.From,
			receiverNumber: cfg.Twillio.To,
			lg:             lg,
		})
	}

	var checker healthy.Checker
	checker.Notifier = notifier
	checker.Logger = lg
	checker.DefaultFailureOptions = &healthy.FailureOptions{
		ReportFailuresCount: cfg.ReportFailuresCount,
		FirstRetryDelay:     time.Duration(cfg.FirstRetryDelay),
	}
	for _, hc := range cfg.HttpChecks {
		check := &healthy.HttpCheck{
			Url:                hc.Url,
			ExpectedStatusCode: hc.ExpectedStatusCode,
			Timeout:            time.Duration(hc.Timeout),
		}
		lg.Printf("Setting up %s, period %s", check.Name(), time.Duration(hc.Period))
		checker.AddTaskWithPeriod(check, time.Duration(hc.Period), time.Duration(hc.Flex))
	}
	lg.Printf("Starting all checks...")
	checker.Run(context.Background())

	stopSignal := make(chan os.Signal)
	signal.Notify(stopSignal, os.Interrupt)
	<-stopSignal
	lg.Printf("Got interrupt signal, shutting down...")
	checker.Stop()
	lg.Printf("Done")
}
