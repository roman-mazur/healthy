package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var twillioClient = &http.Client{Timeout: 5 * time.Second}

type twillio struct {
	accountSid string
	authToken  string

	senderNumber   string
	receiverNumber string

	lg *log.Logger
}

func (t *twillio) Notify(taskName string, e error) {
	formData := &url.Values{}
	formData.Set("To", t.receiverNumber)
	formData.Set("From", t.senderNumber)
	formData.Set("Body", fmt.Sprintf("healthy\nNew feailure detected for %s\n%s", taskName, e))
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSid),
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(t.accountSid, t.authToken)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", "application/json")

	if resp, err := twillioClient.Do(req); err == nil {
		if resp.StatusCode != http.StatusCreated {
			t.lg.Printf("Unexpected Twillio response: %d", resp.StatusCode)
		}
	} else {
		t.lg.Printf("Cannot post to Twillio: %s", err)
	}
}
