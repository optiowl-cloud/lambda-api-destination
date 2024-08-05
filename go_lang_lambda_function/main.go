package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
)

var processedMessageIDs = map[string]struct{}{}

func handler(ctx context.Context, event json.RawMessage) (map[string]interface{}, error) {
	log.Printf("Event: %s", string(event))

	// Decode the JSON event into a slice of SQS messages
	var records []map[string]interface{}
	if err := json.Unmarshal(event, &records); err != nil {
		return map[string]interface{}{
			"statusCode": 400,
			"body":       "Bad Request: Error parsing event",
		}, nil
	}

	// Process each record
	for _, record := range records {
		body, ok := record["body"]
		if !ok {
			log.Printf("Missing 'body' in record: %v", record)
			continue
		}

		// Create a unique hash for the message body to use as a deduplication key
		bodyStr, ok := body.(string)
		if !ok {
			log.Printf("Invalid 'body' type in record: %v", record)
			continue
		}

		messageHash := md5Hash(bodyStr)

		if _, found := processedMessageIDs[messageHash]; found {
			log.Printf("Duplicate message detected: %v", record)
			continue
		}

		processedMessageIDs[messageHash] = struct{}{}

		// Forward the entire SQS message to the external API
		resp, err := forwardToAPI(record)
		if err != nil {
			log.Printf("Request Exception: %s", err)
			continue
		}

		log.Printf("Status Code: %d", resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("Response Body: %s", string(bodyBytes))
	}

	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Request forwarded successfully",
	}, nil
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func forwardToAPI(body interface{}) (*http.Response, error) {
	apiURL := "https://webhook.site/f7274881-d5c1-43f3-9a30-de0a0a255cdd"
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
