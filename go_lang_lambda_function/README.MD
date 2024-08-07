## Setup and run this function

- 1. Get go lange code and Update according to you.
- 2. Run :- `go mod init <your_module_name>`
- 3. Add Dependencies :- `go get github.com/aws/aws-lambda-go/lambda`
- 4. Create Build  :- `GOOS=linux GOARCH=amd64 go build -o bootstrap main.go`
- 5. Create a .zip Package :- `zip deployment.zip bootstrap`
- 6. Now upload this Zip file for code your Lambda

- In this we insert URL with SQS message 
    - example 
    - `{
            "url": "< url >",
            "message_body": "This is a simple text message."
        }`
