package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
)

// Initialize an empty map to keep track of processed message IDs
var processedMessageIDs = make(map[string]struct{})

// Record represents an SQS record
type Record struct {
	MessageID         string                 `json:"messageId"`
	ReceiptHandle     string                 `json:"receiptHandle"`
	Body              string                 `json:"body"`
	Attributes        map[string]interface{} `json:"attributes"`
	MessageAttributes map[string]interface{} `json:"messageAttributes"`
	Md5OfBody         string                 `json:"md5OfBody"`
	EventSource       string                 `json:"eventSource"`
	EventSourceARN    string                 `json:"eventSourceARN"`
	AwsRegion         string                 `json:"awsRegion"`
}

// Event represents the SQS event structure
type Event struct {
	Records []Record `json:"Records"`
}

// Lambda function handler
func handler(event interface{}) (map[string]interface{}, error) {
	var records []Record

	// Determine the type of event and unmarshal accordingly
	switch e := event.(type) {
	case []interface{}:
		// If the event is an array, iterate over each element
		for _, r := range e {
			recordBytes, _ := json.Marshal(r)
			var record Record
			if err := json.Unmarshal(recordBytes, &record); err != nil {
				log.Printf("JSON Decode Error: %v", err)
				continue
			}
			records = append(records, record)
		}
	case map[string]interface{}:
		// If the event is a map with "Records" field, unmarshal into Event
		eventBytes, _ := json.Marshal(e)
		var evt Event
		if err := json.Unmarshal(eventBytes, &evt); err != nil {
			log.Printf("JSON Decode Error: %v", err)
			return nil, err
		}
		records = evt.Records
	default:
		// Return an error if the event format is invalid
		return map[string]interface{}{
			"statusCode": 400,
			"body":       "Bad Request: Invalid event format",
		}, nil
	}

	// Process each record
	for _, record := range records {
		body := record.Body

		// Parse the body content to extract the URL and other data
		var bodyContent map[string]interface{}
		if err := json.Unmarshal([]byte(body), &bodyContent); err != nil {
			log.Printf("JSON Decode Error: %v", err)
			continue
		}

		// Extract the URL from the body content
		URL, ok := bodyContent["url"].(string)
		if !ok {
			log.Printf("Missing 'url' in message body: %v", bodyContent)
			continue
		}
		delete(bodyContent, "url")

		// Create a unique hash for the message body to use as a deduplication key
		messageHash := md5.Sum([]byte(fmt.Sprintf("%v", record)))
		hashString := fmt.Sprintf("%x", messageHash)

		// Check if the message has already been processed
		if _, exists := processedMessageIDs[hashString]; exists {
			log.Printf("Duplicate message detected: %v", record)
			continue
		}

		// Add the message hash to the set of processed messages
		processedMessageIDs[hashString] = struct{}{}

		// Convert the updated bodyContent back to a JSON string
		updatedBody, err := json.Marshal(bodyContent)
		if err != nil {
			log.Printf("JSON Marshal Error: %v", err)
			continue
		}

		// Update the body of the record
		record.Body = string(updatedBody)

		// Marshal the record to JSON
		messageBytes, err := json.Marshal(record)
		if err != nil {
			log.Printf("JSON Marshal Error: %v", err)
			continue
		}

		// Send the updated record to the specified URL
		resp, err := http.Post(URL, "application/json", bytes.NewBuffer(messageBytes))
		if err != nil {
			log.Printf("Request Exception: %v", err)
			continue
		}
		defer resp.Body.Close()

		// Read and log the response
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Status Code: %d", resp.StatusCode)
		log.Printf("Response Body: %s", respBody)
	}

	// Return success response
	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Request forwarded successfully",
	}, nil
}

func main() {
	// Start the Lambda function
	lambda.Start(handler)
}
