#!/usr/bin/env bash

# Deployment script for Lingopaste to Kubernetes
set -e

echo "Deploying Lingopaste to Kubernetes..."

# Create namespace
echo "Creating namespace..."
kubectl apply -f namespace.yaml

# Apply ConfigMap
echo "Applying ConfigMap..."
kubectl apply -f configmap.yaml

# Apply Secrets (ensure secrets.yaml exists and is not the example file)
if [ -f "secrets.yaml" ]; then
    echo "Applying Secrets..."
    kubectl apply -f secrets.yaml
else
    echo "ERROR: secrets.yaml not found!"
    echo "Please create secrets.yaml from secrets.yaml.example"
    exit 1
fi

# Apply Deployments
echo "Deploying applications..."
kubectl apply -f deployment.yaml

# Apply Services
echo "Creating services..."
kubectl apply -f service.yaml

# Apply HPA
echo "Setting up auto-scaling..."
kubectl apply -f hpa.yaml

# Apply Ingress (optional)
if [ -f "ingress.yaml" ]; then
    echo "Configuring ingress..."
    kubectl apply -f ingress.yaml
fi

echo "Deployment complete!"
echo ""
echo "Check status with:"
echo "  kubectl get pods -n lingopaste"
echo "  kubectl get svc -n lingopaste"
echo "  kubectl get hpa -n lingopaste"
