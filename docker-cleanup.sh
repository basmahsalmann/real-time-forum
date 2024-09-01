#!/bin/bash

echo "Cleaning up Docker resources..."

# Remove all stopped containers, unused networks, dangling images, and build cache
echo "Removing stopped containers, unused networks, dangling images, and build cache..."
docker system prune -f

# Remove unused volumes
echo "Removing unused volumes..."
docker volume prune -f

# Remove unused images
echo "Removing unused images..."
docker image prune -f

echo "Docker cleanup complete!"

#To use this script, you would simply need to make it executable and run it:

#chmod +x docker-cleanup.sh
#./docker-cleanup.sh