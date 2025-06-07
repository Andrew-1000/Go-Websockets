#!/bin/bash

echo "Generating self-signed certificate..."    
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout server.key -out server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=example.com"
if [ $? -eq 0 ]; then
    echo "Self-signed certificate generated successfully."
else
    echo "Failed to generate self-signed certificate."
    exit 1
fi
echo "Copying server.key and server.crt to /etc/ssl/private/ and /etc/ssl/certs/ respectively..."
sudo cp server.key /etc/ssl/private/
sudo cp server.crt /etc/ssl/certs/
if [ $? -eq 0 ]; then
    echo "Files copied successfully."
else
    echo "Failed to copy files."
    exit 1
fi
echo "Setting permissions for server.key..."
sudo chmod 600 /etc/ssl/private/server.key
if [ $? -eq 0 ]; then
    echo "Permissions set successfully."
else
    echo "Failed to set permissions."
    exit 1
fi
echo "Setting permissions for server.crt..."
sudo chmod 644 /etc/ssl/certs/server.crt
if [ $? -eq 0 ]; then
    echo "Permissions set successfully."
else
    echo "Failed to set permissions."
    exit 1
fi
echo "Self-signed certificate setup completed successfully."
echo "Restarting web server to apply changes..."
sudo systemctl restart nginx
if [ $? -eq 0 ]; then
    echo "Web server restarted successfully."
else
    echo "Failed to restart web server."
    exit 1
fi
echo "Self-signed certificate generation and setup completed successfully."
echo "You can now use the self-signed certificate for testing purposes."                            