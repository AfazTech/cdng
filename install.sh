#!/bin/bash



# Update package list and install necessary packages
if ! apt update; then
    echo "Failed to update package list."
    exit 1
fi

if ! apt install nginx certbot git -y; then
   echo "Failed to install required packages."
    exit 1
fi

# backup nginx
backup_name="/etc/nginx_backup_$(date +%Y%m%d_%H%M%S)"
mv /etc/nginx "$backup_name"

# Clone the repository into /etc
if ! git clone https://github.com/AfazTech/cdng /tmp/cdng; then
    echo "Failed to clone repository."
    exit 1
fi
mv /tmp/cdng/nginx /etc


if ! systemctl restart nginx; then
    echo  "Failed to restart Nginx."
    exit 1
fi
