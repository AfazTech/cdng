#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Test Nginx configuration
if ! nginx -t; then
    json_response false "Nginx configuration test failed. Reload aborted."
    exit 1
fi

# Reload Nginx
if ! systemctl reload nginx; then
    json_response false "Failed to reload Nginx."
    exit 1
fi

json_response true "Nginx reloaded successfully."