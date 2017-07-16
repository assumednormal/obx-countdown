package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var botID = os.Getenv("BOT_ID")
var port = os.Getenv("PORT")

// Attachment represents a generic attachment in a GroupMe message
type Attachment struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Message represents a single message in a GroupMe chat
type Message struct {
	Attachments []Attachment `json:"attachments"`
	AvatarURL   string       `json:"avatar_url"`
	CreatedAt   int64        `json:"created_at"`
	GroupID     string       `json:"group_id"`
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	SenderID    string       `json:"sender_id"`
	SenderType  string       `json:"sender_type"`
	SourceGUID  string       `json:"source_guid"`
	System      bool         `json:"system"`
	Text        string       `json:"text"`
	UserID      string       `json:"user_id"`
}

func sendCountdown() error {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	end := time.Date(2017, time.August, 27, 13, 0, 0, 0, loc)
	until := end.Sub(time.Now())

	days := int64(until.Hours()) / 24
	hours := int64(until.Hours()) - 24*days
	minutes := int64(until.Minutes()) - 24*60*days - 60*hours
	seconds := int64(until.Seconds()) - 24*60*60*days - 60*60*hours - 60*minutes

	text := fmt.Sprintf("%02d Days %02d Hours %02d Minutes %02d Seconds", days, hours, minutes, seconds)
	body := []byte(fmt.Sprintf("{\"bot_id\": \"%s\", \"text\": \"%s\"}", botID, text))

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

func messageHandler(w http.ResponseWriter, r *http.Request) {
	// must be a POST request
	if r.Method != "POST" {
		return
	}

	// try to decode request body as a Message
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var m Message
	err := decoder.Decode(&m)
	if err != nil {
		return
	}

	// make sure the message includes the cue to respond
	t := strings.ToUpper(m.Text)
	if strings.Contains(t, "COUNTDOWN") || strings.Contains(t, "COUNT DOWN") {
		err = sendCountdown()
		if err != nil {
			return
		}
	}

	return
}

func main() {
	http.HandleFunc("/", messageHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
