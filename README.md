# Spark UI for Kubernetes

This is a simple Spark UI reverse proxy to ease accessing the UI when working with Kubernetes.

## Screenshots

![](docs/screenshot-home.png)

## Architecture

TBC

## Setup
### In Cluster

TBC

### Out of Cluster

TBC

## Development
This project requires :
- skaffold
- kustomize
- kind (for testing on kubernetes cluster)

To install these dependencies, run the following commands
```bash
# macos
brew install skaffold kustomize kind

# linux
apt install skaffold kustomize kind
```

To start modifying the source code, read the [developers guide](test/README.md).
