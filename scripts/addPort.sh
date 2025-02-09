#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Check if port is provided
if [ -z "$1" ]; then
    json_response false "Usage: ./addPort.sh <port>"
    exit 1
fi

PORT=$1

# Validate port using regex
if ! [[ $PORT =~ ^[0-9]{1,5}$ ]]; then
    json_response false "Invalid port format."
    exit 1
fi

# Add the port to the listen.conf file
echo "listen $PORT;" >> /etc/nginx/conf.d/listen.conf

# Reload Nginx using the reload.sh script
RELOAD_OUTPUT=$(./reload.sh)
RELOAD_STATUS=$(echo "$RELOAD_OUTPUT" | jq -r '.ok')

if [[ $RELOAD_STATUS == "true" ]]; then
    json_response true "Port $PORT added successfully."
else
    # If reload fails, remove the domain configuration file
    rm -f /etc/nginx/conf.d/domains/$DOMAIN.conf
    json_response false "$(echo "$RELOAD_OUTPUT" | jq -r '.message')"
fi

