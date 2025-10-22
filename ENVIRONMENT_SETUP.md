# Environment Variables Setup Guide

This document describes all environment variables used in the Voltio CHMMA system and how to configure them.

## Overview

All services in the Voltio CHMMA system use environment variables for configuration instead of hardcoded credentials. This improves security and makes it easier to deploy the system in different environments.

## Configuration Files

- **Root `.env.example`**: Template for consumer services, test producers, and WebSocket server
- **`Backend/automation-engine/.env.example`**: Template specifically for the automation engine service

## Setup Instructions

1. Copy `.env.example` to `.env` in the root directory:
   ```bash
   cp .env.example .env
   ```

2. Copy `Backend/automation-engine/.env.example` to `Backend/automation-engine/.env`:
   ```bash
   cp Backend/automation-engine/.env.example Backend/automation-engine/.env
   ```

3. Edit both `.env` files with your actual configuration values.

4. **IMPORTANT**: Never commit `.env` files to version control. They are already listed in `.gitignore`.

## Environment Variables Reference

### RabbitMQ Configuration

| Variable | Description | Example | Used By |
|----------|-------------|---------|---------|
| `RABBITMQ_URI` | Full connection URI for RabbitMQ | `amqp://user:password@host:5672/` | All services |

### InfluxDB Configuration

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `INFLUX_URL` | InfluxDB server URL | `http://localhost:8086` | Consumer services |
| `INFLUX_TOKEN` | InfluxDB authentication token | (required) | Consumer services |
| `INFLUX_ORG` | InfluxDB organization name | `mi-org` | Consumer services |
| `INFLUX_BUCKET` | InfluxDB bucket name | `sensores` | Consumer services |

### Queue Names

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `ALERTS_QUEUE_NAME` | Queue for alert messages | `alerts-queue` | Consumer services |
| `PZEM_QUEUE_NAME` | Queue for PZEM energy meter data | `PZEM_queue` | PZEM services |
| `DHT22_QUEUE_NAME` | Queue for DHT22 temperature/humidity data | `DHT22_queue` | DHT22 services |
| `LIGHT_QUEUE_NAME` | Queue for light sensor data | `LightSensor_queue` | Light sensor services |
| `PIR_QUEUE_NAME` | Queue for PIR motion sensor data | `PIR_queue` | PIR services |

### WebSocket URIs

| Variable | Description | Used By |
|----------|-------------|---------|
| `PZEM_WEBSOCKET_URI` | WebSocket endpoint for PZEM data | PZEM consumer |
| `DHT22_WEBSOCKET_URI` | WebSocket endpoint for DHT22 data | DHT22 consumer |
| `LIGHT_WEBSOCKET_URI` | WebSocket endpoint for light sensor data | Light sensor consumer |
| `PIR_WEBSOCKET_URI` | WebSocket endpoint for PIR data | PIR consumer |

### Notification Service

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `API_WEBHOOK_URL` | Voltio API webhook endpoint | `http://localhost:8000/api/internal/notifications/service` | Notification consumer |
| `API_WEBHOOK_TOKEN` | Authentication token for webhook | (required) | Notification consumer |

### WebSocket Server

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `WEBSOCKET_PORT` | Port for WebSocket server | `8081` | WebSocket server |

### PostgreSQL Configuration (Automation Engine)

| Variable | Description | Example | Used By |
|----------|-------------|---------|---------|
| `POSTGRES_HOST` | PostgreSQL server hostname | `localhost` | Automation engine |
| `POSTGRES_PORT` | PostgreSQL server port | `5432` | Automation engine |
| `POSTGRES_USER` | Database username | `voltio_user` | Automation engine |
| `POSTGRES_PASSWORD` | Database password | (required) | Automation engine |
| `POSTGRES_DB` | Database name | `voltiodb` | Automation engine |

## Service-Specific Configuration

### Test Producers

All test producers (DHT22, Light, PIR, PZEM) use:
- `RABBITMQ_URI` - Required for publishing test data

If not set, they will default to `amqp://guest:guest@localhost:5672/` for local development.

### Consumer Services

All consumer services require:
- `RABBITMQ_URI` - For consuming messages
- `INFLUX_TOKEN` - For storing metrics
- Service-specific WebSocket URI

If not set, most variables have sensible defaults for local development, except for credentials which should always be provided.

### Automation Engine

Requires all PostgreSQL variables to be set. If `RABBITMQ_URI` is not set, the service will fail to start (no default is provided for security reasons).

## Security Best Practices

1. **Never commit `.env` files** - They contain sensitive credentials
2. **Use strong passwords** - Especially for production environments
3. **Rotate credentials regularly** - Update tokens and passwords periodically
4. **Limit access** - Only provide credentials to services that need them
5. **Use different credentials per environment** - Development, staging, and production should have separate credentials

## Troubleshooting

### Service fails to start with "environment variable required"

This means a required environment variable is not set. Check the error message for which variable is missing and add it to your `.env` file.

### Service connects to wrong server

Check that your `.env` file is in the correct location and that the service is reading from it. Some services may require a restart to pick up environment variable changes.

### InfluxDB authentication fails

Verify that your `INFLUX_TOKEN` is valid and has the necessary permissions for your `INFLUX_ORG` and `INFLUX_BUCKET`.

## Migration from Hardcoded Values

If you're upgrading from a version with hardcoded credentials:

1. Create `.env` files as described above
2. Copy your existing credentials from the old hardcoded values
3. Restart all services to pick up the new environment variables
4. Verify all services are working correctly
5. Delete any backup files containing old credentials

## Support

For issues or questions about environment configuration, please open an issue on the GitHub repository.
