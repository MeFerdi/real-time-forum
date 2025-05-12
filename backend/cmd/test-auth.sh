#!/bin/bash

echo "1. Register new user..."
curl -X POST http://localhost:8080/api/auth/register \
-H "Content-Type: application/json" \
-d '{
    "nickname": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User",
    "age": 25,
    "gender": "male"
}'

sleep 2

echo -e "\n\n2. Login with new user..."
LOGIN_RESPONSE=$(curl -X POST http://localhost:8080/api/auth/login \
-H "Content-Type: application/json" \
-d '{
    "identifier": "test@example.com",
    "password": "password123"
}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo -e "\nAuth Token: $TOKEN"

# Save token for later use
echo $TOKEN > .auth_token