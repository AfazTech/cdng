#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Check if domain is provided
if [ -z "$1" ]; then
    json_response false "Usage: ./getSSL.sh <domain>"
    exit 1
fi

DOMAIN=$1

# Obtain SSL certificate using Certbot
if ! certbot certonly --nginx -d $DOMAIN --non-interactive --agree-tos -m admin@$DOMAIN; then
    json_response false "Failed to obtain SSL certificate for $DOMAIN."
    exit 1
fi

json_response true "SSL certificate for $DOMAIN obtained successfully."