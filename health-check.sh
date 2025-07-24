#!/bin/bash

# Voltio Services Health Check Script
# This script verifies that all services are running correctly

echo "🔍 Checking Voltio Services Health..."

# Check if Docker is running
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ Docker daemon is not running"
    exit 1
fi

echo "✅ Docker is running"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed or not in PATH"
    exit 1
fi

echo "✅ Docker Compose is available"

# Check if services are running
echo "📊 Checking service status..."
if ! docker-compose ps | grep -q "voltio-all-services"; then
    echo "❌ Voltio services container is not running"
    echo "💡 Try: docker-compose up -d"
    exit 1
fi

# Check container health
HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' voltio-all-services 2>/dev/null || echo "no-health")

if [ "$HEALTH_STATUS" = "healthy" ]; then
    echo "✅ Container health check: HEALTHY"
elif [ "$HEALTH_STATUS" = "unhealthy" ]; then
    echo "⚠️ Container health check: UNHEALTHY"
    echo "📋 Recent health check logs:"
    docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' voltio-all-services | tail -5
elif [ "$HEALTH_STATUS" = "starting" ]; then
    echo "🟡 Container health check: STARTING (waiting for services to initialize)"
else
    echo "🔶 Container health check: NOT CONFIGURED"
fi

# Check WebSocket server
echo "🌐 Testing WebSocket server..."
if curl -f -s http://localhost:8081/health &> /dev/null; then
    echo "✅ WebSocket server is responding"
else
    echo "❌ WebSocket server is not responding on port 8081"
fi

# Check individual services in container
echo "🔍 Checking individual services..."
SUPERVISOR_STATUS=$(docker-compose exec -T voltio-services supervisorctl status 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "$SUPERVISOR_STATUS" | while read line; do
        service_name=$(echo $line | awk '{print $1}')
        service_status=$(echo $line | awk '{print $2}')
        
        if [[ "$service_status" == "RUNNING" ]]; then
            echo "✅ $service_name: RUNNING"
        else
            echo "❌ $service_name: $service_status"
        fi
    done
else
    echo "⚠️ Cannot check individual services (container might be starting)"
fi

# Check logs for errors
echo "📋 Checking recent logs for errors..."
ERROR_COUNT=$(docker-compose logs --tail=50 2>/dev/null | grep -i "error\|failed\|panic" | wc -l)

if [ "$ERROR_COUNT" -eq 0 ]; then
    echo "✅ No recent errors found in logs"
else
    echo "⚠️ Found $ERROR_COUNT potential errors in recent logs"
    echo "💡 Check logs with: docker-compose logs"
fi

# Check environment configuration
echo "⚙️ Checking environment configuration..."
if [ -f ".env" ]; then
    echo "✅ Environment file (.env) exists"
    
    # Check critical variables
    if grep -q "RABBITMQ_URI" .env && grep -q "INFLUX_URL" .env; then
        echo "✅ Critical environment variables are configured"
    else
        echo "⚠️ Some critical environment variables might be missing"
        echo "💡 Check your .env file configuration"
    fi
else
    echo "⚠️ Environment file (.env) not found"
    echo "💡 Copy .env.example to .env and configure it"
fi

# Summary
echo ""
echo "📊 Health Check Summary:"
echo "========================"

# Overall status
if docker-compose ps | grep -q "Up.*healthy"; then
    echo "🟢 Overall Status: HEALTHY"
    echo "🌐 WebSocket Server: http://localhost:8081"
    echo "📊 View logs: docker-compose logs -f"
    echo "📈 Service status: docker-compose exec voltio-services supervisorctl status"
elif docker-compose ps | grep -q "Up"; then
    echo "🟡 Overall Status: RUNNING (health check pending)"
    echo "💡 Wait a moment and run this script again"
else
    echo "🔴 Overall Status: NOT RUNNING"
    echo "💡 Start services with: docker-compose up -d"
fi

echo ""
echo "🚀 For more detailed monitoring:"
echo "   docker-compose logs -f                    # View all logs"
echo "   docker-compose exec voltio-services bash  # Access container shell"
echo "   docker stats                              # Resource usage"
