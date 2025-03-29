package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Arlan-Z/def-betting-api/internal/domain" // local import, to remember
)

type NotificationHandler struct {
	// Add dependencies here if needed (e.g., database connection, message queue producer)
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) HandleEventNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" && !startsWith(contentType, "application/json;") {
		log.Printf("WARN: Received notification with potentially incorrect Content-Type: %s", contentType)
	}

	var notification domain.EventNotification
	err := json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		log.Printf("ERROR: Failed to decode notification body: %v", err)
		http.Error(w, "Bad request body", http.StatusBadRequest)
		return
	}
	// It's good practice to close the body
	defer r.Body.Close()

	// --- Process the notification ---
	// TODO:
	// - Update your internal database state for the event.
	// - Trigger downstream actions (e.g., payout calculations, user notifications).
	// - Potentially send the data to a message queue.
	log.Printf("INFO: Received event notification:")
	log.Printf("  EventID: %s", notification.EventID)
	log.Printf("  EventName: %s", notification.EventName)
	log.Printf("  Result: %s", notification.Result)
	// --- End processing ---

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notification received successfully.")
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
