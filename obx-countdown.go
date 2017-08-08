package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GroupMe Bot ID
var botID = os.Getenv("BOT_ID")

// Twilio Info
var accountSID = os.Getenv("ACCOUNT_SID")
var authToken = os.Getenv("AUTH_TOKEN")
var from = os.Getenv("FROM")

// Port asssigned by Heroku
var port = os.Getenv("PORT")

// date of vacation and time zone - see init()
var end time.Time
var loc time.Location

// gif to include in message
var gif = os.Getenv("GIF")

// GroupMeMessage represents a single message in a GroupMe chat
type GroupMeMessage struct {
	Text string `json:"text"`
}

func timeRemainingText() string {
	until := end.Sub(time.Now())

	days := int64(until.Hours()) / 24
	hours := int64(until.Hours()) - 24*days
	minutes := int64(until.Minutes()) - 24*60*days - 60*hours
	seconds := int64(until.Seconds()) - 24*60*60*days - 60*60*hours - 60*minutes

	return fmt.Sprintf("%02d Days %02d Hours %02d Minutes %02d Seconds", days, hours, minutes, seconds)
}

func groupMeSendCountdown() error {
	text := timeRemainingText()
	body := []byte(fmt.Sprintf("{\"bot_id\": \"%s\", \"text\": \"%s\", \"attachments\": [{\"type\": \"image\", \"url\": \"%s\"}]}", botID, text, gif))

	r, err := http.NewRequest("POST", "https://api.groupme.com/v3/bots/post", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if _, err := client.Do(r); err != nil {
		return err
	}

	return nil
}

func groupMeHandler(w http.ResponseWriter, r *http.Request) {
	// must be a POST request
	if r.Method != "POST" {
		return
	}

	// try to decode request body as a message
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var m GroupMeMessage
	err := decoder.Decode(&m)
	if err != nil {
		return
	}

	// make sure the message includes the cue to respond
	t := strings.ToUpper(m.Text)
	if strings.Contains(t, "COUNTDOWN") || strings.Contains(t, "COUNT DOWN") {
		err = groupMeSendCountdown()
		if err != nil {
			return
		}
	}

	return
}

func twilioSendCountdown(to string) error {
	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Body", timeRemainingText())

	r, err := http.NewRequest("POST",
		"https://api.twilio.com/2010-04-01/Accounts/"+accountSID+"/Messages.json",
		bytes.NewBuffer([]byte(v.Encode())))
	if err != nil {
		return err
	}

	r.SetBasicAuth(accountSID, authToken)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	if _, err := client.Do(r); err != nil {
		return err
	}

	return nil
}

func twilioHandler(w http.ResponseWriter, r *http.Request) {
	// must be a POST Request
	if r.Method != "POST" {
		return
	}

	// make sure the message includes the cue to respond
	t := strings.ToUpper(r.FormValue("Body"))
	if strings.Contains(t, "COUNTDOWN") || strings.Contains(t, "COUNT DOWN") {
		err := twilioSendCountdown(r.FormValue("From"))
		if err != nil {
			return
		}
	}

	return
}

func init() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}

	end = time.Date(2017, time.August, 27, 13, 0, 0, 0, loc)
}

func main() {
	http.HandleFunc("/groupme", groupMeHandler)
	http.HandleFunc("/twilio", twilioHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
