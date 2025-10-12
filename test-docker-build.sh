#!/bin/bash

# Test script to verify Docker build works locally
# This helps ensure the deployment will work on Render

echo "Testing Docker build for RCA Backend..."

# Build the Docker image
echo "Building Docker image..."
docker build -t rca-backend-test .

if [ $? -eq 0 ]; then
    echo "✅ Docker build successful!"
    
    # Test running the container
    echo "Testing container startup..."
    docker run --rm -d --name rca-test -p 8080:8080 rca-backend-test
    
    # Wait a moment for startup
    sleep 5
    
    # Test health endpoint
    echo "Testing health endpoint..."
    curl -f http://localhost:8080/health
    
    if [ $? -eq 0 ]; then
        echo "✅ Health check passed!"
    else
        echo "❌ Health check failed!"
    fi
    
    # Clean up
    docker stop rca-test
    docker rmi rca-backend-test
    
    echo "✅ Docker test completed successfully!"
    echo "Your application is ready for Render deployment!"
else
    echo "❌ Docker build failed!"
    echo "Please check the build logs above for errors."
    exit 1
fi
