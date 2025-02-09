#!/bin/bash

# Function to return JSON response
json_response() {
    local status=$1
    local message=$2
    jq -n --argjson ok "$status" --arg message "$message" '{ok: $ok, message: $message}'
}

# Get the number of active domains
DOMAIN_COUNT=$(ls /etc/nginx/conf.d/domains/ | wc -l)

# Get total upload and download traffic using vnstat
TRAFFIC=$(vnstat --json | jq '.interfaces[] | {upload: .traffic[0].tx, download: .traffic[0].rx}')

# Get CPU load (1-minute average)
CPU_LOAD=$(uptime | awk -F 'load average:' '{print $2}' | awk -F, '{print $1}' | xargs)

# Get CPU usage percentage
CPU_PERCENT=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *$[0-9.]*$%* id.*/\1/" | awk '{print 100 - $1}')

# Get RAM usage percentage
RAM_PERCENT=$(free | grep Mem | awk '{print ($3/$2) * 100}')

# Display the information in JSON format
jq -n \
--argjson domain_count "$DOMAIN_COUNT" \
--argjson traffic "$TRAFFIC" \
--arg cpu_load "$CPU_LOAD" \
--argjson cpu_percent "$CPU_PERCENT" \
--argjson ram_percent "$RAM_PERCENT" \
'{
  ok: true,
  domain_count: $domain_count,
  traffic: $traffic,
  cpu_load: $cpu_load,
  cpu_percent: $cpu_percent,
  ram_percent: $ram_percent
}'
