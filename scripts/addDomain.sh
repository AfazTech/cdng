#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Check if domain and IP are provided
if [ -z "$1" ] || [ -z "$2" ]; then
    json_response false "Usage: ./addDomain.sh <domain> <ip>"
    exit 1
fi

DOMAIN=$1
IP=$2

# Validate domain using regex
if ! [[ $DOMAIN =~ ^([a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]\.)+[a-zA-Z]{2,}$ ]]; then
    json_response false "Invalid domain format."
    exit 1
fi

# Validate IP using regex
if ! [[ $IP =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
    json_response false "Invalid IP format."
    exit 1
fi

# request free ssl
RELOAD_OUTPUT=$(./getSSL.sh)
RELOAD_STATUS=$(echo "$RELOAD_OUTPUT" | jq -r '.ok')

if [[ $RELOAD_STATUS == "true" ]]; then
    json_response true "certbot request ssl successfully."
else
    # If certbot fails, 
    json_response false "$(echo "$RELOAD_OUTPUT" | jq -r '.message')"
    exit 1
fi


# Create the domain configuration file
cat > /etc/nginx/conf.d/domains/$DOMAIN.conf <<EOF
server {
    include /etc/nginx/conf.d/listen.conf;
    server_name $DOMAIN;

    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;

    location / {
        proxy_pass \$scheme://$IP:\$server_port;
    }
}
EOF

# Reload Nginx using the reload.sh script
RELOAD_OUTPUT=$(./reload.sh)
RELOAD_STATUS=$(echo "$RELOAD_OUTPUT" | jq -r '.ok')

if [[ $RELOAD_STATUS == "true" ]]; then
    json_response true "Domain $DOMAIN added successfully."
else
    # If reload fails, remove the domain configuration file
    rm -f /etc/nginx/conf.d/domains/$DOMAIN.conf
    json_response false "$(echo "$RELOAD_OUTPUT" | jq -r '.message')"
fi