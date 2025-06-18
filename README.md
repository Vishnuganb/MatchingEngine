# MatchingEngine

## Overview
The `MatchingEngine` is a high-performance order matching engine designed for financial trading systems. 
It supports order book management, trade execution, and integration with external systems like RabbitMQ and Kafka for real-time messaging.

## Features

- ⚡ **Order Matching**: Supports limit orders with full and partial fills.
- 🔁 **Event Handling**: Emits events for order lifecycle stages—new, executed, partially filled, canceled, and rejected.
- 🛢️ **Database Integration**: Uses PostgreSQL for persisting orders.
- 📬 **Messaging Queues**:
  - Accepts incoming orders via RabbitMQ.
  - Publishes execution reports via Kafka.
- 🧱 **Modular Architecture**: Clean separation of concerns for handler, service, repository, and messaging layers.
- 📈 **Scalable Design**: Built for high-throughput and low-latency trading applications.
---

## Prerequisites

Ensure the following tools are installed:

- [Docker](https://www.docker.com/)
- [Go 1.21+](https://go.dev/doc/install)
- [migrate CLI](https://github.com/golang-migrate/migrate#installation)
- [sqlc](https://docs.sqlc.dev/en/latest/)
- [goimports-reviser](https://github.com/incu6us/goimports-reviser)
- [golangci-lint](https://golangci-lint.run/)

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
make up 
```

### 3. Create the DB:
```bash
  make createdb
```

### 3. Apply Databse Migrations:
```bash
  make migrateup
```

## Makefile Commands

The `Makefile` in this project provides several commands to simplify common tasks. Below is an explanation of each command:

### Database Commands
- **`createdb`**: Creates the `orderManagement` database inside the PostgreSQL container.
- **`dropdb`**: Drops the `orderManagement` database if it exists.

### Migration Commands
- **`migrateup`**: Applies all database migrations to bring the schema up to date.
- **`migratedown`**: Rolls back the applied database migrations.

### Code Generation
- **`sqlc`**: Generates Go code for database queries using `sqlc`.

### Testing
- **`test`**: Runs all tests in the project with verbose output and code coverage.

## 🧰 Using Makefile

This project includes a `Makefile` to simplify common development tasks.

### 🔧 Commands

- Start the service with Docker:
  ```bash
  make up
  ```
  
- Stop the service with Docker:
  ```bash
  make down
  ```

- Create the `orderManagement` database:
  ```bash
  make createdb
  ```

- Drop the database (if exists):
  ```bash
  make dropdb
  ```

- Apply database migrations:
  ```bash
  make migrateup
  ```

- Roll back the last migration:
  ```bash
  make migratedown
  ```

- Generate Go code from SQL queries using `sqlc`:
  ```bash
  make sqlc
  ```

- Format and revise imports:
  ```bash
  make import-reviser
  ```

- Run lint checks:
  ```bash
  make lint
  ```

- Run unit-tests with coverage:
  ```bash
  make unit-test
  ```
  
- Run component tests:
  ```bash
  cd MatchingEngine/integration
  make integration-test
  ```

