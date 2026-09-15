# HealthConnect - Storage Service (Go)

The **Storage Service** is a high-performance microservice built in Go (Fiber), designed to manage binary data storage, static asset delivery, zero-trust server-side security, and asynchronous event-driven file replication within the **HealthConnect** ecosystem.

---

## 🛡️ Architecture & Security Features

* **Zero-Trust File Storage:** Eliminates frontend exposure of storage API keys by acting as an isolated internal vault.
* **Storage Strategy Pattern:** Supports modular providers (`LOCAL` file system and extensible to cloud providers like `AWS S3`).
* **At-Rest Encryption:** Optional AES-GCM envelope encryption per bucket before writing bytes to disk.
* **Low-Memory Footprint:** Uses non-blocking streams (`io.Reader` and `io.Writer`) and Fiber's native `SendFile` to serve binary payloads with zero-copy overhead.
* **Asynchronous Replication:** Integrates with **RabbitMQ** to handle background file synchronization across buckets without blocking client HTTP responses.

---

## 🚀 Tech Stack

* **Language:** Go 1.21+
* **Framework:** Fiber v2
* **ORM:** GORM (PostgreSQL Driver)
* **Message Broker:** RabbitMQ (AMQP 0-9-1)
* **Database:** PostgreSQL 14+

---

## ⚙️ Environment Variables (.env)

```env
# Server Configuration
PORT=8083
FRONTEND_ORIGINS=http://localhost:5174

# PostgreSQL Database
DB_HOST=inventory-postgres
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=storage_db
DB_PORT=5432
DB_SSLMODE=disable

# Security & API Authentication
API_KEY=your_internal_api_key
API_SECRET=your_internal_api_secret
STORAGE_CIPHER_KEY=12345678901234567890123456789012 # Must be exactly 32 bytes

# RabbitMQ Event Broker
RABBITMQ_URL=amqp://guest:guest@localhost:5672/