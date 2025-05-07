# Game Deployment Guide (Dockerized)

This document provides step-by-step instructions to build, run, and play the game using Docker.

---

## 1. Pre-requisites

Before deploying the game, ensure the following are installed:

- Docker: https://www.docker.com/get-started
- Git: https://git-scm.com/downloads
- (Optional) Docker Compose: https://docs.docker.com/compose/ — not required for this version

Verify installation:

```bash
docker --version
git --version
```

---

## 2. Clone the Repository

Open your terminal and run:

```bash
git clone https://github.com/yourusername/your-repo-name.git
cd your-repo-name
```

---

## 3. Build the Docker Image

Run the following command in the root project directory (where the Dockerfile is located):

```bash
docker build -t go-code-breaker .
```

---

## 4. Run the Docker Container

Start the game server:

```bash
docker run -p 8080:8080 go-code-breaker
```

This maps port 8080 of the container to your local machine.

---

## 5. Connect and Play the Game

The game is command-line based. To connect as a client:

1. Open a new terminal window
2. Ensure Go is installed on your system
3. Run the client:

```bash
go run GO/main.go client
```

This connects to the server at localhost:8080.

---

End of Guide