#!/bin/bash

echo "🚀 Starting Neuron VPS Initialization..."

# 1. Create the temporary runner directory (Fixes 'bind source path does not exist')
echo "📂 Creating runner temp directory..."
mkdir -p /var/www/neuron/backend/tmp/runner
# Ensure current user owns it
sudo chown -R $USER:$USER /var/www/neuron/backend/tmp/runner
echo "✅ Directory created."

# 2. Pull all required Docker Images
echo "🐳 Pulling Docker images..."

# Python
echo "   - Pulling Python..."
docker pull python:3.12-alpine

# C++ (GCC)
echo "   - Pulling GCC..."
docker pull gcc:latest

# Java (Example - based on common registry patterns)
echo "   - Pulling OpenJDK..."
docker pull openjdk:17-alpine

# Node.js (Example)
echo "   - Pulling Node..."
docker pull node:18-alpine

# Go (Example)
echo "   - Pulling Go..."
docker pull golang:1.22-alpine

echo "🎉 Initialization Complete! You can now run the worker."