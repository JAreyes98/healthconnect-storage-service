# HealthConnect - Storage Service (Go)

The **Storage Service** is a high-performance microservice built in Go, designed to handle large-scale binary data storage with a focus on security and low latency. It serves as the second security vault in the **HealthConnect** ecosystem.

## 🛡️ Second-Layer Security & Storage

This service implements an "Encrypted at Rest" strategy with a modular architecture:

1.  **Double Encryption (Layer 2):** Even though data arrives encrypted from the Java service, this service applies an additional encryption layer using provider-specific configurations before persisting the data.
2.  **Strategy Pattern:** Built with a flexible storage strategy, allowing seamless switching between:
    * **Local File System:** For development and on-premise deployments.
    * **AWS S3 / Cloud Storage:** For scalable production environments.
3.  **Atomic Metadata Management:** Ensures that file metadata (UUIDs, Physical Paths, and Content Types) is synchronized with the physical storage via GORM and PostgreSQL.

## 🚀 Technology Stack

* **Go (Golang) 1.21+**
* **Fiber v2:** An Express-inspired web framework for high performance.
* **GORM:** ORM for secure and efficient PostgreSQL interaction.
* **Afero:** A filesystem abstraction layer for testing and multi-storage support.
* **Google UUID:** For generating unique, non-collision physical file identifiers.

## 🏗️ Technical Highlights

* **Streaming Buffers:** Utilizes `io.Reader` and `io.Writer` streams to process files without loading them entirely into memory, minimizing the RAM footprint even with large medical images.
* **Concurrency:** Built to handle multiple simultaneous uploads/downloads leveraging Go's lightweight goroutines.
* **Obfuscation:** Decouples original filenames from physical storage names using UUID-based mapping to prevent metadata leakage at the OS level.

## 🛠️ Configuration & Installation

### Prerequisites
* Go 1.21 or higher
* PostgreSQL 14+

### Environment Variables (.env)
```env
DB_URL=postgres://user:pass@localhost:5432/storage_db
API_KEY=your_internal_api_key
API_SECRET=your_internal_api_secret
STORAGE_CIPHER_KEY=#32 characters key
DB_HOST=your_db_host
DB_USER=your_postgres_user
DB_PASSWORD=your_password
DB_NAME=storage_db
DB_PORT=5432
DB_SSLMODE=disable


```bash
docker run -d  --name healthconnect-storage-service -p 8082:8082 --network healthconnect-net --env-file .env healthconnect-strorage-service:1.0.0