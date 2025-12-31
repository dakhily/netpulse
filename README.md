# Netpulse
> **High-concurrency network observability prober built in Go.**

[![Status: Work in Progress](https://img.shields.io/badge/status-work--in--progress-orange)](#)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/dakhily/netpulse)](https://goreportcard.com/report/github.com/dakhily/netpulse)

Netpulse is a production-grade network health monitoring tool. It bridges the gap between traditional "up/down" checks and deep SRE observability by tracking **p99 latency histograms**, jitter, and error rates across multiple endpoints simultaneously.

## Tech Stack
- **Language:** Go (High-concurrency Goroutines)
- **Observability:** Prometheus & Grafana
- **Infrastructure:** Docker & Docker Compose
- **Cloud (Future):** AWS Multi-region deployment

---

## Architecture
Netpulse runs as a containerized service that probes targets and exposes a `/metrics` endpoint. Prometheus scrapes these metrics, and Grafana visualizes them to provide real-time insights into network performance.



---

## Getting Started

### Prerequisites
- Docker & Docker Compose installed.
- Go 1.21+ (if running locally).

### Installation & Setup
1. **Clone the repository:**
   ```bash
   git clone [https://github.com/dakhily/netpulse.git](https://github.com/dakhily/netpulse.git)
   cd netpulse
2. **Spin up the stack (Prober + Prometheus + Grafana):**
   ```bash
   docker compose -f 'docker-compose.yml' up -d --build
3. **Verify the health:**
    - Prober Metrics: http://localhost:8080/metrics
    - Prometheus: http://localhost:9090
    - Grafana: http://localhost:5000

4. **Stopping the System:**
   ```bash
   docker compose down

License

Distributed under the GPLv3 License. See LICENSE for more information.

Maintained by Dakhil Y.