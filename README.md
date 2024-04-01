# Spark UI for Kubernetes

![CICD](https://github.com/helkaroui/spark-k8s-reverse-proxy/actions/workflows/release.yml/badge.svg)
![Docker Image Version](https://img.shields.io/docker/v/helkaroui/spark-reverse-proxy)
![Docker Pulls](https://img.shields.io/docker/pulls/helkaroui/spark-reverse-proxy)

This is a simple Spark UI reverse proxy to ease accessing the UI when working with Kubernetes.

## Screenshots

Home Page :
![](docs/screenshot-home.png)

Driver's Logs Page :
![](docs/screenshot-logs.png)

Driver's Manifest Page :
![](docs/screenshot-manifest.png)

## Architecture
The reverse proxy is deployed as an application inside the kubernetes cluster and can route traffic from an ingress to the spark driver ui.

![](docs/diagram.jpg)

## Setup
The reverse proxy relies on label selection to list spark drivers, thus you need to add the following label depending on
submission mode :
- In client mode: you need to add the label `spark-role=driver` and expose port 4040
- In cluster mode: all labels are already added by default

In the spark submit command, you need to enable reverse proxy as follows :
```bash
/opt/spark/bin/./spark-submit \
    --master k8s://https://kubernetes.default.svc:443 \
    --deploy-mode client \
    --name $JOB_NAME \
    ...\
    --conf spark.ui.reverseProxy=true \
    file:///opt/spark/examples/jars/spark-examples_2.12-3.5.0.jar "$1"
```

## Usage
### Docker
TBC
```
docker pull helkaroui/spark-reverse-proxy:latest
```

### Kubectl
TBC


### Helm
To install using Helm chart :
```bash
# Clone the repository then run :
helm install my-release ./helm
```

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
