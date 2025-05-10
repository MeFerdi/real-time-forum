#!/bin/bash

# Login and capture token
LOGIN_RESPONSE=$(curl -X POST http://localhost:8080/api/auth/login \
-H "Content-Type: application/json" \
-d '{
    "identifier": "test@example.com",
    "password": "password123"
}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

echo -e "\n3. Create new post..."
curl -X POST http://localhost:8080/api/posts/create \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{
    "title": "My First Post",
    "content": "This is a test post content",
    "categories": ["Technology"]
}'

echo -e "\n4. List all posts..."
curl -X GET "http://localhost:8080/api/posts?page=1" \
-H "Authorization: Bearer $TOKEN"

echo -e "\n5. Get post details..."
curl -X GET http://localhost:8080/api/posts/1 \
-H "Authorization: Bearer $TOKEN"

echo -e "\n6. Add comment to post..."
curl -X POST http://localhost:8080/api/posts/1/comments \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{
    "content": "This is my first comment!"
}'

echo -e "\n7. View post with comments..."
curl -X GET http://localhost:8080/api/posts/1 \
-H "Authorization: Bearer $TOKEN"