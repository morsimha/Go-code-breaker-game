# Kubernetes Deployment Guide

This guide explains how to deploy the game application to a Kubernetes cluster.

---

## 1. Pre-requisites

Make sure the following are installed on your machine:

- Docker
- kubectl (Kubernetes CLI)
- Access to a Kubernetes cluster (e.g., Minikube, Docker Desktop with Kubernetes enabled)

Verify installation:

```bash
docker --version
kubectl version --client
```

---

## 2. Build and Push Docker Images

The Kubernetes cluster must have access to the required Docker images. You can either:

### Option A: Push to Docker Hub

```bash
docker build -t your-dockerhub-username/go-code-breaker .
docker push your-dockerhub-username/go-code-breaker

docker build -f Dockerfile.client -t your-dockerhub-username/go-game-client .
docker push your-dockerhub-username/go-game-client
```

Then update `k8s-deployment.yaml` to use these image names.

### Option B: Use Minikube’s Docker Daemon

```bash
eval $(minikube docker-env)
docker build -t go-code-breaker .
docker build -f Dockerfile.client -t go-game-client .
```

---

## 3. Apply Kubernetes Configurations

Run the following command:

```bash
kubectl apply -f k8s-deployment.yaml
```

This will:

- Deploy the game server and client
- Create a service to expose the server

---

## 4. Access the Game Server

Check the external port:

```bash
kubectl get svc game-server-service
```

If using Minikube:

```bash
minikube service game-server-service
```

---

## 5. Interact with the Client

Get the client pod name:

```bash
kubectl get pods
```

Then run:

```bash
kubectl logs <game-client-pod>
kubectl exec -it <game-client-pod> -- sh
```

---

End of Guide