#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Update package list and install necessary packages
if ! apt update; then
    json_response false "Failed to update package list."
    exit 1
fi

if ! apt install nginx certbot jq git vnstat -y; then
    json_response false "Failed to install required packages."
    exit 1
fi

# backup nginx
mv /etc/nginx /etc/nginx_backup

# Clone the repository into /etc
if ! git clone https://github.com/imafaz/cdng /etc/nginx; then
    json_response false "Failed to clone repository."
    exit 1
fi

# Test Nginx configuration
if ! nginx -t; then
    json_response false "Nginx configuration test failed."
    exit 1
fi

# Restart Nginx to apply changes
if ! systemctl restart nginx; then
    json_response false "Failed to restart Nginx."
    exit 1
fi

json_response true "Nginx and required packages installed and configured."