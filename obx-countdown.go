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

// GroupMe Bot ID
var botID = os.Getenv("BOT_ID")

// Port asssigned by Heroku
var port = os.Getenv("PORT")

// date of vacation and time zone - see init()
var end time.Time
var loc time.Location

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

func groupMeSendCountdown(isError bool) error {
	text := ""
	if isError {
		text = "incorrect format"
	} else {
		text = timeRemainingText()
	}
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
	if (strings.Contains(t, "COUNTDOWN") || strings.Contains(t, "COUNT DOWN")) && strings.Contains(t, "SET") {
		str := strings.Split(t, " ")[2]
		layout := "2006-01-02T15:04:05"
		t, err := time.ParseInLocation(layout, str, &loc)
		if err != nil {
			err2 := groupMeSendCountdown(true)
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		end = t
		err = groupMeSendCountdown(false)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.Contains(t, "COUNTDOWN") || strings.Contains(t, "COUNT DOWN") {
		err = groupMeSendCountdown(false)
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

	end = time.Date(2018, time.September, 1, 13, 0, 0, 0, loc)
}

func main() {
	http.HandleFunc("/groupme", groupMeHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
