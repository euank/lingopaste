# Deployment Guide

## GitHub Container Registry Setup

### Initial Configuration (One-time)

1. **Push code to GitHub**
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin git@github.com:YOUR_USERNAME/lingopaste.com.git
   git push -u origin main
   ```

2. **Enable GitHub Actions**
   - Go to your repository on GitHub
   - Navigate to Settings → Actions → General
   - Ensure "Allow all actions and reusable workflows" is selected

3. **Package Visibility** (Optional)
   - After first build, packages will appear at `https://github.com/YOUR_USERNAME?tab=packages`
   - Click on each package (backend/frontend)
   - Go to Package settings
   - Change visibility to Public (if desired)

### Automated Builds

Images are automatically built and pushed on:
- **Every push to `main`** → Tagged as `latest` + git SHA
- **Every push to `develop`** → Tagged with branch name + git SHA
- **Pull requests to `main`** → Build only (no push)
- **Manual trigger** → Via "Actions" tab → "Build All Images" → "Run workflow"

### Image Locations

After building, images will be available at:
```
ghcr.io/YOUR_USERNAME/lingopaste.com/backend:latest
ghcr.io/YOUR_USERNAME/lingopaste.com/backend:main-abc1234
ghcr.io/YOUR_USERNAME/lingopaste.com/frontend:latest
ghcr.io/YOUR_USERNAME/lingopaste.com/frontend:main-abc1234
```

## Using the Images

### Pull Images (from your K8s cluster or server)

1. **Authenticate with GitHub Container Registry**
   ```bash
   # Create a GitHub Personal Access Token (PAT)
   # Go to: GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
   # Create new token with permissions: read:packages
   
   # Login to ghcr.io
   echo "YOUR_GITHUB_PAT" | docker login ghcr.io -u YOUR_USERNAME --password-stdin
   ```

2. **Pull images**
   ```bash
   docker pull ghcr.io/YOUR_USERNAME/lingopaste.com/backend:latest
   docker pull ghcr.io/YOUR_USERNAME/lingopaste.com/frontend:latest
   ```

### Update Kubernetes Deployments

1. **Update image references in K8s manifests**
   
   Edit `k8s/deployment.yaml`:
   ```yaml
   # Replace:
   image: lingopaste-backend:latest
   # With:
   image: ghcr.io/YOUR_USERNAME/lingopaste.com/backend:latest
   
   # Replace:
   image: lingopaste-frontend:latest
   # With:
   image: ghcr.io/YOUR_USERNAME/lingopaste.com/frontend:latest
   ```

2. **Create K8s image pull secret** (if packages are private)
   ```bash
   kubectl create secret docker-registry ghcr-login-secret \
     --docker-server=ghcr.io \
     --docker-username=YOUR_USERNAME \
     --docker-password=YOUR_GITHUB_PAT \
     --docker-email=YOUR_EMAIL \
     -n lingopaste
   ```

3. **Reference secret in deployment**
   
   Add to `k8s/deployment.yaml`:
   ```yaml
   spec:
     imagePullSecrets:
     - name: ghcr-login-secret
     containers:
     - name: backend
       image: ghcr.io/YOUR_USERNAME/lingopaste.com/backend:latest
   ```

4. **Deploy**
   ```bash
   cd k8s
   ./deploy.sh
   ```

## Continuous Deployment (Future)

To enable automatic deployment to K8s after successful builds:

1. Add K8s cluster credentials to GitHub Secrets
2. Create deploy workflow that runs after build
3. Use `kubectl set image` or ArgoCD/Flux for GitOps

Example manual deploy after build:
```bash
# After GitHub Actions builds new images
kubectl set image deployment/lingopaste-backend \
  backend=ghcr.io/YOUR_USERNAME/lingopaste.com/backend:main-$(git rev-parse --short HEAD) \
  -n lingopaste

kubectl set image deployment/lingopaste-frontend \
  frontend=ghcr.io/YOUR_USERNAME/lingopaste.com/frontend:main-$(git rev-parse --short HEAD) \
  -n lingopaste
```

## Monitoring Builds

- View build status: `https://github.com/YOUR_USERNAME/lingopaste.com/actions`
- View packages: `https://github.com/YOUR_USERNAME?tab=packages`
- Build logs available for 90 days

## Troubleshooting

**Build fails with permission error:**
- Ensure GitHub Actions has write permissions to packages
- Go to Settings → Actions → General → Workflow permissions
- Select "Read and write permissions"

**Can't pull images:**
- Check if packages are private (need authentication)
- Verify GitHub PAT has `read:packages` scope
- Ensure docker login was successful

**Want to delete old images:**
- Go to package page
- Click on specific version
- Delete version (keeps latest)
