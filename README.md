<div align="center">

# GeoStream

### Distributed IoT Telemetry & Anomaly Detection Platform

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org)
[![Python](https://img.shields.io/badge/Python-3.9+-3776AB?style=flat&logo=python&logoColor=white)](https://python.org)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react&logoColor=black)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![Kafka](https://img.shields.io/badge/Apache-Kafka-231F20?style=flat&logo=apache-kafka&logoColor=white)](https://kafka.apache.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**A production-ready microservices platform for real-time fleet management and telemetry analytics**

</div>

---

## 🎯 Overview

**GeoStream** is an enterprise-grade, event-driven microservices platform designed to handle massive-scale IoT telemetry data from fleet vehicles. Built with Go microservices, Apache Kafka event streaming, and real-time geospatial analytics, it demonstrates modern distributed systems architecture patterns.

This project showcases:
- **Event-driven architecture** with Kafka pub/sub
- **Hybrid database strategy** (SQL + NoSQL + Cache)
- **gRPC** for internal microservice communication
- **WebSocket** real-time updates to frontend
- **Geospatial indexing** with Redis
- **Containerized deployment** with Docker Compose

---

## 🚨 Problem Statement

### The Challenge

Logistics and supply chain companies managing fleets of **10,000+ vehicles** face critical "Big Data" challenges:

| Challenge | Impact | Requirement |
|-----------|--------|-------------|
| **Volume** | Terabytes of sensor data daily (GPS, temp, speed, fuel) | Need Kafka for high-throughput ingestion |
| **Latency** | Query 10,000 truck locations instantly | Need Redis geo-spatial indexing |
| **Storage** | Archive years of route history | Need time-series storage (PostgreSQL) |
| **Detection** | Detect crashes/failures in milliseconds | Need gRPC for low-latency validation |

### The Solution

GeoStream provides a **scalable, distributed platform** that:
1. **Ingests** sensor data from thousands of devices simultaneously
2. **Processes** data through event-driven microservices
3. **Detects** anomalies in real-time (e.g., engine temp > 100°C)
4. **Visualizes** live vehicle positions on an interactive map
5. **Archives** historical data for compliance and analytics

---

## 🏗️ System Architecture

```
                    ┌─────────────────────────────────────────┐
                    │   IoT Device Simulator (5000 vehicles)  │
                    │      Python Script (Go Routines)        │
                    └────────────────┬────────────────────────┘
                                     │ Protobuf messages (every 2s)
                                     ▼
                    ┌─────────────────────────────────────────┐
                    │       Ingestion Layer (Go Backend)       │
                    │    HTTP/WebSocket Server (Port 8080)     │
                    └────────────────┬────────────────────────┘
                                     │ Kafka Producer
                                     ▼
                    ┌─────────────────────────────────────────┐
                    │         Apache Kafka (Port 9092)         │
                    │       Topic: vehicle-telemetry          │
                    └──┬──────────────┬───────────────────┬───┘
                       │              │                   │
          ┌────────────▼──┐  ┌───────▼────────┐  ┌──────▼─────────┐
          │  Consumer A    │  │  Consumer B     │  │  Consumer C     │
          │  Real-Time     │  │  Archiver       │  │  Inspector      │
          │  Processor     │  │  Service        │  │  (gRPC:50051)   │
          └────┬───────────┘  └───────┬─────────┘  └────────────────┘
               │                      │
               │ Update               │ Store
               ▼                      ▼
    ┌──────────────────┐   ┌─────────────────────┐
    │   Redis Cache    │   │   PostgreSQL DB     │
    │  (Port 6379)     │   │   (Port 5432)       │
    │  Geo-Index       │   │   Vehicle Metadata  │
    └────┬─────────────┘   └─────────────────────┘
         │ WebSocket Push
         ▼
    ┌──────────────────────────┐
    │   React Frontend (UI)    │
    │   Leaflet.js Map         │
    │   Live Vehicle Tracking  │
    └──────────────────────────┘
```

### Component Breakdown

#### 1. **Data Source (Simulation)**
- **Role**: Simulates 5,000 trucks sending telemetry every 2 seconds
- **Tech**: Python script with concurrent requests
- **Location**: `/scripts/iot_simulator.py`

#### 2. **Ingestion Layer**
- **Role**: High-throughput HTTP/WebSocket endpoint for sensor data
- **Tech**: Go (Fiber framework), Kafka Producer
- **Location**: `/backend/` (main API server)
- **Ports**: 8080 (HTTP), 50051 (gRPC)

#### 3. **Message Broker**
- **Role**: Decouples producers and consumers for scalability
- **Tech**: Apache Kafka 3.7.0 (KRaft mode - no Zookeeper)
- **Topics**: `vehicle-telemetry`, `anomaly-alerts`
- **Port**: 9092

#### 4. **Processing Microservices**

| Service | Role | Technology | Output |
|---------|------|------------|--------|
| **Real-Time Consumer** | Updates live vehicle positions | Go + Redis GEO | WebSocket to frontend |
| **Archiver Consumer** | Stores historical telemetry | Go + PostgreSQL | Time-series data |
| **Inspector (gRPC)** | Validates anomalies before alerts | Go gRPC | Alert triggers |

#### 5. **Storage Layer**

| Database | Type | Purpose | Port |
|----------|------|---------|------|
| **PostgreSQL** | Relational (SQL) | Driver profiles, vehicle metadata, org charts | 5432 |
| **Redis** | In-memory cache | Geo-spatial index, real-time locations | 6379 |

#### 6. **Frontend Dashboard**
- **Role**: Live map showing vehicle positions
- **Tech**: React + TypeScript, Leaflet.js
- **Features**: Real-time WebSocket updates, anomaly notifications
- **Location**: `/frontend/ui/`

---

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.21+ (Goroutines for concurrency)
- **Web library**: net/http
- **Protocols**: REST, WebSocket, gRPC
- **Message Queue**: Apache Kafka 3.7.0

### Frontend
- **Framework**: Vite React 18+ with TypeScript
- **Mapping**: Leaflet.js
- **Styling**: CSS Modules / Tailwind CSS
- **Real-time**: WebSocket client

### Data Stores
- **PostgreSQL 15**: Relational data (drivers, vehicles)
- **Redis 7**: Geo-spatial indexing (GEORADIUS commands)

### DevOps
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Monitoring**: Kafka UI (http://localhost:8085)

### Protocols & Standards
- **gRPC**: Low-latency inter-service communication
- **Protocol Buffers**: Service definitions (`/proto/`)
- **WebSocket**: Real-time frontend updates

---

## ✨ Features

### Core Capabilities
- ✅ **High-Throughput Ingestion**: Handle 5,000+ concurrent IoT streams
- ✅ **Real-Time Geospatial Queries**: Query vehicle locations within 10km radius in <10ms
- ✅ **Event-Driven Processing**: Kafka-based decoupling for horizontal scalability
- ✅ **Anomaly Detection**: gRPC-powered rule engine (e.g., engine temp > 100°C)
- ✅ **Live Dashboard**: WebSocket-powered map showing moving vehicles
- ✅ **Historical Analysis**: Time-series data for route playback
- ✅ **Production-Ready**: Health checks, graceful shutdown, structured logging

### Why This Architecture Stands Out
1. **Hybrid Database Strategy**: Demonstrates knowing when to use SQL vs NoSQL vs Cache
2. **Event-Driven Design**: Not just CRUD - implements true pub/sub patterns
3. **Protocol Versatility**: Uses REST (external), gRPC (internal), WebSocket (real-time)
4. **Concurrency Mastery**: Go goroutines handle thousands of simultaneous connections
5. **Production Patterns**: Health checks, circuit breakers, retry logic, monitoring

---

## 🚀 Quick Start

### Prerequisites

Ensure you have the following installed:

| Tool | Version | Check Command |
|------|---------|---------------|
| **Docker** | 20.10+ | `docker --version` |
| **Docker Compose** | 2.0+ | `docker-compose --version` |
| **Python** | 3.9+ | `python3 --version` |
| **Go** | 1.21+ (for local dev) | `go version` |

**Hardware Requirements**: Minimum 8GB RAM available for containers

### Installation Steps

#### 1. Clone the Repository
```bash
git clone https://github.com/THETITAN220/GeoStream.git
cd GeoStream
```

#### 2. Start Infrastructure Services
```bash
# Start all services (Postgres, Redis, Kafka, Backend)
docker-compose up -d

# Verify all services are healthy
docker-compose ps
```

**Expected Output**:
```
NAME                    STATUS              PORTS
geostream-db            Up (healthy)        5432
geostream-redis         Up (healthy)        6379
geostream-kafka         Up (healthy)        9092
geostream-kafka-ui      Up                  8085
geostream-backend       Up (healthy)        8080, 50051
```

#### 3. Access Service Dashboards

| Service | URL | Purpose |
|---------|-----|---------|
| **Kafka UI** | http://localhost:8085 | Monitor topics, consumer groups |
| **Backend API** | http://localhost:8080 | REST endpoints |
| **gRPC Inspector** | localhost:50051 | Anomaly validation service |
| **Frontend** | http://localhost:3000 | Live vehicle map (if running) |

#### 4. Run the IoT Simulator

The simulator generates telemetry for 5,000 virtual vehicles:

```bash
# Navigate to scripts directory
cd scripts

# Install Python dependencies (one-time setup)
pip install -r requirements.txt

# Run the simulator
python3 simulator.py
```

**Expected Output**:
```
[INFO] Starting IoT Simulator - 5000 vehicles
[SUCCESS] ✓ Vehicle-0001: lat=37.7749, lon=-122.4194, speed=65km/h, temp=85°C
[SUCCESS] ✓ Vehicle-0002: lat=40.7128, lon=-74.0060, speed=72km/h, temp=78°C
...
[INFO] Sent 5000 telemetry updates in 1.2s
```

#### 5. Verify Data Flow

**Check Kafka Messages**:
1. Open Kafka UI: http://localhost:8085
2. Navigate to **Topics** → `truck-telemetry`
3. View incoming messages in real-time

### Stopping Services

```bash
# Stop all containers
docker-compose down

# Stop and remove volumes (fresh start)
docker-compose down -v
```

---
## ⚙️ Configuration

### Environment Variables

All services use environment variables defined in `docker-compose.yaml` but below is a `.env` reference:

```yaml
# --- PROJECT SETTINGS ---
PROJECT_NAME=geostream

# --- DATABASE CREDENTIALS ---
DB_USER=geo
DB_PASSWORD=geostream
DB_NAME=geostream

# --- KAFKA CONFIGURATION (KRaft Mode) ---
KAFKA_CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk

# --- PORTS ---
DB_PORT=5432
REDIS_PORT=6379
KAFKA_PORT=9092
UI_PORT=8085
APP_PORT=8080
GRPC_PORT=50051

```
---

## 👨‍💻 Development

### Local Development Setup

#### 1. Install Go Dependencies
```bash
go mod download
go mod tidy
```

#### 2. Generate gRPC Code from Proto Files
```bash
# Install protoc compiler
# macOS: brew install protobuf
# Linux: apt-get install protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

#### 3. Run Backend Locally (requires Docker services)
```bash
# Start only infrastructure (Postgres, Redis, Kafka)
docker-compose up -d postgres redis kafka

# Run Go backend on host machine
cd backend
go run main.go
```

#### 4. Run Frontend Locally
```bash
cd frontend/ui
npm install
npm run dev
# Open http://localhost:3000
```
---

## 🚢 Deployment

### Production Deployment Options

#### Option 1: AWS ECS (Elastic Container Service)

1. **Containerize Services**: Already done via `Dockerfile`
2. **Push to ECR** (Elastic Container Registry):
   ```bash
   aws ecr get-login-password --region us-west-2 | \
     docker login --username AWS --password-stdin <account>.dkr.ecr.us-west-2.amazonaws.com
   
   docker tag geostream-backend:latest <account>.dkr.ecr.us-west-2.amazonaws.com/geostream:latest
   docker push <account>.dkr.ecr.us-west-2.amazonaws.com/geostream:latest
   ```
3. **Create ECS Task Definitions** for each service
4. **Deploy to ECS Cluster**

#### Option 2: Kubernetes (EKS or self-hosted)

Convert `docker-compose.yaml` to Kubernetes manifests:

```yaml
# deployment.yaml (example for backend)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: geostream-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: geostream-backend
  template:
    spec:
      containers:
      - name: backend
        image: geostream-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: KAFKA_BROKERS
          value: "kafka-service:9092"
```

### Managed Services (Production-Ready)

Replace self-hosted components with managed alternatives:

| Self-Hosted | AWS Managed | Benefits |
|-------------|-------------|----------|
| Kafka | **AWS MSK** (Managed Streaming for Kafka) | Auto-scaling, monitoring |
| Redis | **AWS ElastiCache** (Redis) | High availability, backups |
| PostgreSQL | **AWS RDS** (PostgreSQL) | Automated backups, read replicas |
| Secrets | **.env files** | **AWS Secrets Manager** | Encrypted credential rotation |

**Example MSK Configuration**:
```bash
# Update environment variable
KAFKA_BROKERS=b-1.msk-cluster.kafka.us-west-2.amazonaws.com:9092
```

### CI/CD Pipeline (GitHub Actions)

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy GeoStream
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker Image
        run: docker build -t geostream-backend backend/
      
      - name: Push to ECR
        run: |
          aws ecr get-login-password | docker login ...
          docker push ...
      
      - name: Deploy to ECS
        run: aws ecs update-service --cluster geostream --service backend --force-new-deployment
```

---
