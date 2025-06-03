# MatchingEngine

## Overview
The `MatchingEngine` is a high-performance order matching engine designed for financial trading systems. 
It supports order book management, trade execution, and integration with external systems like RabbitMQ and Kafka for real-time messaging.

## Features
- **Order Matching**: Supports buy and sell orders with price and quantity matching.
- **Event Handling**: Generates events for new orders, fills, partial fills, cancellations, and rejections.
- **Database Integration**: Uses PostgreSQL for persistent storage of orders and events.
- **Message Queues**: Integrates with RabbitMQ for order requests and Kafka for event notifications.
- **Scalable Architecture**: Designed to handle high-throughput trading scenarios.

---

## Prerequisites
- **Docker**: Ensure Docker is installed for running services like PostgreSQL, RabbitMQ, and Kafka.
- **Go**: Requires Go 1.24 or later.
- **Migrate**: Install the `migrate` tool for database migrations.

---

## Setup Instructions

### 1. Clone the Repository
```bash
git clone https://github.com/Vishnuganb/MatchingEngine.git
cd MatchingEngine 
```

### 2. Run the Application
```bash
cd integration
docker-compose up -d
```

### 3. Run the Go application:
```bash
  go run cmd/server/main.go
```

## Makefile Commands

The `Makefile` in this project provides several commands to simplify common tasks. Below is an explanation of each command:

### Database Commands
- **`postgres`**: Starts a PostgreSQL container for the Matching Engine.
- **`createdb`**: Creates the `orderManagement` database inside the PostgreSQL container.
- **`dropdb`**: Drops the `orderManagement` database if it exists.

### Migration Commands
- **`migrateup`**: Applies all database migrations to bring the schema up to date.
- **`migratedown`**: Rolls back the applied database migrations.

### Code Generation
- **`sqlc`**: Generates Go code for database queries using `sqlc`.

### Testing
- **`test`**: Runs all tests in the project with verbose output and code coverage.

### Usage
Run any of the commands using:
```bash
make <command>
