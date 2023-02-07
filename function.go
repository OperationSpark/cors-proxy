package apiprox

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/errorreporting"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/go-chi/chi/v5"
)

func init() {
	// Register an HTTP function with the Functions Framework
	// This handler name maps to the entry point name in the Google Cloud Function platform.
	// https://cloud.google.com/functions/docs/writing/write-http-functions
	functions.HTTP("CloudFunc", newApp().ServeHTTP)

	// Disable log prefixes such as the default timestamp.
	// Prefix text prevents the message from being parsed as JSON.
	// A timestamp is added when shipping logs to Cloud Logging.
	log.SetFlags(0)
}

var errorClient *errorreporting.Client

func newApp() *chi.Mux {
	// Avoid variable shadow for errorClient
	var err error
	errorClient, err = initErrorReporting()
	if err != nil {
		log.Fatalf("initErrorReporting: %v", err)
	}

	return New().Router
}

func initErrorReporting() (*errorreporting.Client, error) {
	if os.Getenv("CI") == "true" {
		return &errorreporting.Client{}, nil
	}

	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	errorClient, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName: "cors-proxy",
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		return nil, err
	}
	return errorClient, nil
}
