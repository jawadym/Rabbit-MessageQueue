#!/bin/bash
# make sure that rabbitmq is running
echo "Message Queue status:"
sudo docker-compose up -d
echo "--------------------------"

# build go binary
echo "Building new executable..."
/usr/local/go/bin/go build || echo "BUILD FAILED"
echo "--------------------------"
