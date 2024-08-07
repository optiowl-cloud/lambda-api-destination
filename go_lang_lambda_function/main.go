package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

// Lambda function handler for raw events
func handler(event json.RawMessage) (map[string]interface{}, error) {
	// Get the URL from the environment variable
	forwardURL := os.Getenv("FORWARD_URL")
	if forwardURL == "" {
		log.Printf("Environment variable FORWARD_URL not set")
		return map[string]interface{}{
			"statusCode": 500,
			"body":       "Internal Server Error: FORWARD_URL not set",
		}, nil
	}

	// Log the entire received event data
	log.Printf("Raw Event Data: %s", event)

	// Send the raw event data to the specified URL
	resp, err := http.Post(forwardURL, "application/json", bytes.NewBuffer(event))
	if err != nil {
		log.Printf("Request Exception: %v", err)
		return map[string]interface{}{
			"statusCode": 500,
			"body":       "Internal Server Error: Request exception",
		}, nil
	}
	defer resp.Body.Close()

	// Read and log the response
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Status Code: %d", resp.StatusCode)
	log.Printf("Response Body: %s", respBody)

	// Return success response
	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Event forwarded successfully",
	}, nil
}

func main() {
	// Start the Lambda function
	lambda.Start(handler)
}
