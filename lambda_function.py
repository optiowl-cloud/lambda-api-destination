import requests
import json
import hashlib

# Initialize an empty set to keep track of processed message IDs
processed_message_ids = set()

def lambda_handler(event, context):
    # Log the event object for debugging
    print(f"Event: {json.dumps(event)}")

    # Check if the event is a list
    if isinstance(event, list):
        records = event
    elif isinstance(event, dict) and 'Records' in event:
        records = event['Records']
    else:
        return {
            'statusCode': 400,
            'body': 'Bad Request: No Records found in event'
        }

    # Extract message body from the SQS event
    for record in records:
        # Ensure the record has a 'body' field
        if 'body' not in record:
            print(f"Missing 'body' in record: {json.dumps(record)}")
            continue

        try:
            # Create a unique hash for the message body to use as a deduplication key
            message_hash = hashlib.md5(json.dumps(record, sort_keys=True).encode('utf-8')).hexdigest()
            
            # Check if the message has already been processed
            if message_hash in processed_message_ids:
                print(f"Duplicate message detected: {json.dumps(record)}")
                continue
            
            # Add the message hash to the set of processed messages
            processed_message_ids.add(message_hash)
            
            # Forward the entire SQS message to the external API
            response = requests.post('https://webhook.site/f7274881-d5c1-43f3-9a30-de0a0a255cdd', json=record)
            
            # Log response for debugging
            print(f"Status Code: {response.status_code}")
            print(f"Response Body: {response.text}")

        except requests.exceptions.RequestException as e:
            # Log any request exceptions
            print(f"Request Exception: {str(e)}")
        except json.JSONDecodeError as e:
            # Log any JSON decode exceptions
            print(f"JSON Decode Error: {str(e)}")

    return {
        'statusCode': 200,
        'body': 'Request forwarded successfully'
    }
