package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	err := SendAlert("bot-testing", "", "Yep")
	log.Println(err)
}

type SlackResponse struct {
	Ok               bool             `json:"ok"`
	Error            string           `json:"error"`
	Warning          string           `json:"warning"`
	ResponseMetadata ResponseMetadata `json:"response_metadata"`
}

type ResponseMetadata struct {
	Warnings []string `json:"warnings"`
}

type Payload struct {
	Channel     string       `json:"channel,omitempty"`
	UserName    string       `json:"username,omitempty"`
	Text        string       `json:"text,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	AsUser      bool         `json:"as_user,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type AttachmentAction struct {
	Name  string `json:"name,omitempty"`
	Text  string `json:"text,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	URL   string `json:"url,omitempty"`
}

type Attachment struct {
	Text           string             `json:"text,omitempty"`
	Fallback       string             `json:"fallback,omitempty"`
	Color          string             `json:"color,omitempty"`
	AttachmentType string             `json:"attachment_type,omitempty"`
	CallbackID     string             `json:"callback_id,omitempty"`
	Actions        []AttachmentAction `json:"actions,omitempty"`
}

func SendAlert(channel string, status string, text string) error {
	data := Payload{
		Channel:  channel,
		UserName: "slacks",
		// Text:        "Some Payload Text",
		// IconURL:     "https://slack.global.ssl.fastly.net/9fa2/img/services/hubot_128.png",
		IconEmoji: statusEmoji(status),
		AsUser:    false,
		Attachments: []Attachment{{
			Text:  text,
			Color: statusColor(status),
			//AttachmentType: "button",
			//Actions:        []AttachmentAction{{
			//	Name:  "Button",
			//	Text:  "Action",
			//	Type:  "button",
			//	Value: "accept",
			//}},
		}},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", os.ExpandEnv("Bearer ${SLACK_TOKEN}"))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	fmt.Println("Status:", resp.Status)
	defer resp.Body.Close()
	var slackResp SlackResponse
	err = json.NewDecoder(resp.Body).Decode(&slackResp)
	if err != nil {
		return err
	}
	// 200 OK does not necessarily mean it worked.
	// {
	//  "ok":false,
	//  "error":"token_revoked",
	//  "warning":"missing_charset",
	//  "response_metadata": {"warnings":["missing_charset"]}
	// }
	if !slackResp.Ok {
		return errors.New(slackResp.Error)
	}
	return nil
}

func statusEmoji(status string) string {
	switch status {
	case "SUCCESS":
		return ":white_check_mark:"
	case "FAILURE", "CANCELLED":
		return ":x:"
	case "STATUS_UNKNOWN", "INTERNAL_ERROR":
		return ":interrobang:"
	default:
		return ":question:"
	}
}

func statusColor(status string) string {
	switch status {
	case "SUCCESS":
		return "good"
	case "FAILURE", "CANCELLED", "STATUS_UNKNOWN", "INTERNAL_ERROR":
		return "danger"
	default:
		return "warning"
	}
}
