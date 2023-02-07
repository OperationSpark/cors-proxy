package main

import (
	"log"
	"net/http"
	"os"

	"github.com/operationspark/apiprox"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9876"
	}

	log.Printf("server starting on port: %s...", port)
	if err := http.ListenAndServe(":"+port, apiprox.New().Router); err != nil {
		log.Fatal(err)
	}
}
