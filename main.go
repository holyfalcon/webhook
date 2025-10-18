package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const verifyToken = "1234"

type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Contacts         []Contact `json:"contacts"`
	Messages         []Message `json:"messages"`
}

type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type Contact struct {
	Profile Profile `json:"profile"`
	WaId    string  `json:"wa_id"`
}

type Profile struct {
	Name string `json:"name"`
}

type Message struct {
	From      string `json:"from"`
	Id        string `json:"id"`
	Timestamp int64  `json:"timestamp,string"`
	Text      Text   `json:"text,omitempty"`
	Type      string `json:"type"`
}

type Text struct {
	Body string `json:"body"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Service is healthy")
	})

	http.HandleFunc("/webhook", webhookHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		getWebhookHandler(w, r)
	} else if r.Method == http.MethodPost {
		postWebhookHandler(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getWebhookHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == verifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return
}

func postWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(payload.Entry) > 0 && len(payload.Entry[0].Changes) > 0 {

		metadata := payload.Entry[0].Changes[0].Value.Metadata
		profileName := payload.Entry[0].Changes[0].Value.Contacts[0].Profile.Name
		profileWaId := payload.Entry[0].Changes[0].Value.Contacts[0].WaId
		receiveMessageFrom := payload.Entry[0].Changes[0].Value.Messages[0].From
		MessageTextBody := payload.Entry[0].Changes[0].Value.Messages[0].Text.Body

		fmt.Printf("Display Phone: %s\n", metadata.DisplayPhoneNumber)
		fmt.Printf("Phone Number ID: %s\n", metadata.PhoneNumberID)

		fmt.Println("Profile Name: ", profileName)
		fmt.Println("Profile WaId: ", profileWaId)
		fmt.Println("Receive Message From: ", receiveMessageFrom)
		fmt.Println("Receive Message Text: ", MessageTextBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"display_phone_number": metadata.DisplayPhoneNumber,
			"phone_number_id":      metadata.PhoneNumberID,
		})
	} else {
		http.Error(w, "No data found", http.StatusBadRequest)
		return
	}
}
