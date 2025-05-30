#!/bin/bash

# Biamp Microservice API Test Script
# This script tests all endpoints from the Biamp Microservice Postman collection

# Configuration variables - Update these values as needed
MICROSERVICE_URL="localhost:8080"
DEVICE_FQDN="biamp-device.local"
INSTANCE_TAG="main"
PRESET_ID="1"

echo "Starting Biamp Microservice API Tests..."
echo "Microservice URL: $MICROSERVICE_URL"
echo "Device FQDN: $DEVICE_FQDN"
echo "Instance Tag: $INSTANCE_TAG"
echo "=============================================="

# GET Volume
echo "Testing GET Volume..."
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/volume/$INSTANCE_TAG/1"
sleep 1

# GET Audiomute
echo "Testing GET Audiomute..."
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomute/$INSTANCE_TAG/1"
sleep 1

# GET Voicelift
echo "Testing GET Voicelift..."
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/voicelift"
sleep 1

# GET Logicselector
echo "Testing GET Logicselector..."
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/logicselector/$INSTANCE_TAG"
sleep 1

# GET Audiomode
echo "Testing GET Audiomode..."
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomode/$INSTANCE_TAG"
sleep 1

echo "=============================================="
echo "Starting SET/PUT operations..."
echo "=============================================="

# SET Volume
echo "Testing SET Volume (25)..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/volume/$INSTANCE_TAG/1" \
     -H "Content-Type: application/json" \
     -d "\"25\""
sleep 1

# SET Audiomute
echo "Testing SET Audiomute (true)..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomute/$INSTANCE_TAG/1" \
     -H "Content-Type: application/json" \
     -d "\"true\""
sleep 1

# SET Preset
echo "Testing SET Preset..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/preset/$PRESET_ID" \
     -H "Content-Type: application/json" \
     -d ""
sleep 1

# SET Voicelift
echo "Testing SET Voicelift (true)..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/voicelift" \
     -H "Content-Type: application/json" \
     -d "\"true\""
sleep 1

# SET Logicselector
echo "Testing SET Logicselector (1)..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/logicselector/$INSTANCE_TAG" \
     -H "Content-Type: application/json" \
     -d "\"1\""
sleep 1

# SET Audiomode
echo "Testing SET Audiomode (1)..."
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomode/$INSTANCE_TAG" \
     -H "Content-Type: application/json" \
     -d "\"1\""
sleep 1

echo "=============================================="
echo "All API tests completed!"
echo "=============================================="
