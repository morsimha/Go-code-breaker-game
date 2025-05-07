# Game Deployment Guide (Docker + Docker Compose)

This document provides complete instructions to build, run, and deploy the game using Docker and Docker Compose.

---

## 1. Pre-requisites

Make sure the following software is installed:

- Docker: https://www.docker.com/get-started
- Git: https://git-scm.com/downloads
- (Optional) Docker Compose: https://docs.docker.com/compose/

Check versions:

```bash
docker --version
git --version
```

---

## 2. Clone the Repository

```bash
git clone https://github.com/yourusername/your-repo-name.git
cd your-repo-name
```

---

## 3. Dockerfile Setup

Ensure the following files exist:

- `Dockerfile` for the game server (produces binary `mygame`)
- `Dockerfile.client` for the client (produces binary `myclient`)

---

## 4. Run with Docker Compose

To run the full game environment (server and client):

```bash
docker-compose up --build
```

This will:
- Build both the server and client images
- Run the server on port 8080
- Start a client container that connects to the server

---

## 5. Interact with the Game

The game is command-line based. To view client output:

```bash
docker logs game-client
```

To attach directly to the client terminal:

```bash
docker attach game-client
```

To stop all containers:

```bash
docker-compose down
```

---

## 6. Optional: Manual Docker Commands

To build and run the server manually:

```bash
docker build -t go-code-breaker -f Dockerfile .
docker run -p 8080:8080 go-code-breaker
```

To build and run the client manually:

```bash
docker build -t go-game-client -f Dockerfile.client .
docker run --network=host go-game-client
```

---

End of Guide