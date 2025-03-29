package main

// VERY HARD PART
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arlan-Z/def-betting-api/config"
	"github.com/Arlan-Z/def-betting-api/internal/client"
	"github.com/Arlan-Z/def-betting-api/internal/handler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		log.Println("INFO: Received shutdown signal, shutting down...")
		cancel()
	}()

	cfg := config.LoadConfig()

	// Create external API client
	apiClient, err := client.NewClient(cfg.ExternalAPIbaseURL)
	if err != nil {
		log.Fatalf("FATAL: Failed to create API client: %v", err)
	}

	// --- Example Client Usage (Optional - for demonstration) ---
	// You might run this section conditionally or remove it for a pure listener service
	runClientExamples(ctx, apiClient, cfg.CallbackBaseURL)
	// --- End Example Client Usage ---

	// Setup HTTP server for receiving notifications
	notificationHandler := handler.NewNotificationHandler()
	mux := http.NewServeMux()
	// The path here should match the path part of your configured CALLBACK_BASE_URL
	// e.g. if CALLBACK_BASE_URL="http://my.service.com/hooks/betting", the path is "/hooks/betting"
	// For simplicity, we assume the callback URL is just the base + /notify
	mux.HandleFunc("/notify", notificationHandler.HandleEventNotification)

	server := &http.Server{
		Addr:    ":" + cfg.ListenPort,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("INFO: Starting callback listener on port %s", cfg.ListenPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Could not listen on %s: %v\n", cfg.ListenPort, err)
		}
	}()

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()

	// Perform graceful shutdown
	log.Println("INFO: Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("FATAL: Server graceful shutdown failed: %v", err)
	}
	log.Println("INFO: Server stopped gracefully.")
}

// runClientExamples demonstrates how to use the API client.
func runClientExamples(ctx context.Context, apiClient *client.Client, callbackBaseURL string) {
	log.Println("--- Running API Client Examples ---")

	// 1. Get All Events
	log.Println("Attempting to get all events...")
	events, err := apiClient.GetAllEvents(ctx)
	if err != nil {
		log.Printf("ERROR getting all events: %v", err)
	} else if len(events) > 0 {
		log.Printf("Successfully retrieved %d events. First event ID: %s, Name: %s", len(events), events[0].ID, events[0].EventName)

		eventIDToUse := events[0].ID // Use the ID of the first event for other examples

		// 2. Get Specific Event
		log.Printf("Attempting to get event %s...", eventIDToUse)
		event, err := apiClient.GetEvent(ctx, eventIDToUse)
		if err != nil {
			log.Printf("ERROR getting event %s: %v", eventIDToUse, err)
		} else {
			log.Printf("Successfully retrieved event: %+v", event)
		}

		// 3. Get Event Details
		log.Printf("Attempting to get details for event %s...", eventIDToUse)
		details, err := apiClient.GetEventDetails(ctx, eventIDToUse)
		if err != nil {
			log.Printf("ERROR getting event details %s: %v", eventIDToUse, err)
		} else {
			log.Printf("Successfully retrieved event details: %+v", details)
		}

		// 4. Subscribe to Event
		// IMPORTANT: The callback URL must be publicly accessible to where *this*
		// service is running, including the correct path ("/notify").
		myCallbackURL := fmt.Sprintf("%s/notify", callbackBaseURL)
		log.Printf("Attempting to subscribe to event %s with callback %s...", eventIDToUse, myCallbackURL)
		subResp, err := apiClient.SubscribeToEvent(ctx, eventIDToUse, myCallbackURL)
		if err != nil {
			log.Printf("ERROR subscribing to event %s: %v", eventIDToUse, err)
		} else {
			log.Printf("Successfully subscribed: %s", subResp)
		}

	} else {
		log.Println("No events returned or error occurred.")
	}

	log.Println("--- Finished API Client Examples ---")
}
