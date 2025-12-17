# 🚀 Jenkins CI/CD Pipeline Setup Guide

Complete guide untuk setup Jenkins CI/CD pipeline untuk deploy Go backend dari GitHub ke Docker.

---

## 📋 Prerequisites

### 1. Jenkins Running in Docker
```bash
# Verify Jenkins is running
docker ps | grep jenkins
```

### 2. Docker Network Exists
```bash
# Check if network exists
docker network ls | grep grafana-mysql-network

# Create if not exists
docker network create grafana-mysql-network
```

### 3. Jenkins Has Docker Access
Jenkins container harus bisa akses Docker daemon.

---

## 🔧 Step 1: Configure Jenkins Container

### Option A: If Jenkins Already Running

**Give Jenkins Docker Access:**

```bash
# Find Jenkins container ID
JENKINS_CONTAINER=$(docker ps | grep jenkins | awk '{print $1}')

# Stop Jenkins
docker stop $JENKINS_CONTAINER

# Run Jenkins with Docker socket access
docker run -d \
  --name jenkins \
  --network grafana-mysql-network \
  -p 8088:8080 \
  -p 50000:50000 \
  -v jenkins_home:/var/jenkins_home \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -u root \
  jenkins/jenkins:lts
```

### Option B: Fresh Jenkins Installation

```bash
# Create Jenkins with Docker access
docker run -d \
  --name jenkins \
  --network grafana-mysql-network \
  -p 8088:8080 \
  -p 50000:50000 \
  -v jenkins_home:/var/jenkins_home \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -u root \
  jenkins/jenkins:lts

# Wait for Jenkins to start
sleep 30

# Get initial admin password
docker exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword
```

---

## 🔧 Step 2: Install Required Jenkins Plugins

### 2.1 Access Jenkins Web UI

Open: `http://localhost:8088`

### 2.2 Install Plugins

Go to: **Manage Jenkins** → **Manage Plugins** → **Available**

Install these plugins:
- ✅ **Docker Pipeline**
- ✅ **Git**
- ✅ **Pipeline**
- ✅ **GitHub Integration**
- ✅ **Workspace Cleanup**

**Or via CLI:**

```bash
docker exec jenkins jenkins-plugin-cli --plugins \
  docker-workflow \
  git \
  workflow-aggregator \
  github \
  ws-cleanup
```

---

## 🔧 Step 3: Install Docker in Jenkins Container

```bash
# Enter Jenkins container
docker exec -it -u root jenkins bash

# Install Docker CLI
apt-get update
apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get install -y docker-ce-cli

# Verify Docker works
docker --version
docker ps

# Exit container
exit
```

---

## 🔧 Step 4: Create Jenkins Pipeline Job

### 4.1 Create New Pipeline

1. Click **"New Item"**
2. Enter name: `merchant-backend-deploy`
3. Select **"Pipeline"**
4. Click **OK**

### 4.2 Configure Pipeline

**General Section:**
- ✅ Check "GitHub project"
- Project url: `https://github.com/aristit0/merchant_summary_demo_backend`

**Build Triggers:**
- ✅ Check "Poll SCM"
- Schedule: `H/5 * * * *` (check every 5 minutes)

**Pipeline Section:**
- Definition: **"Pipeline script from SCM"**
- SCM: **Git**
- Repository URL: `https://github.com/aristit0/merchant_summary_demo_backend.git`
- Branch: `*/main`
- Script Path: `Jenkinsfile`

**OR use "Pipeline script" and paste the Jenkinsfile content directly**

### 4.3 Save Configuration

Click **"Save"**

---

## 🔧 Step 5: Add Jenkinsfile to Repository

### Option A: If You Own The Repo

```bash
# Clone repo
git clone https://github.com/aristit0/merchant_summary_demo_backend.git
cd merchant_summary_demo_backend

# Add Jenkinsfile
# (copy Jenkinsfile content dari file yang saya berikan)

# Add Dockerfile
# (copy Dockerfile content dari file yang saya berikan)

# Commit and push
git add Jenkinsfile Dockerfile
git commit -m "Add Jenkins CI/CD pipeline"
git push origin main
```

### Option B: If You Don't Own The Repo

**Use "Pipeline script" directly in Jenkins:**
- Copy entire Jenkinsfile content
- Paste into Jenkins Pipeline script section
- Save

---

## 🚀 Step 6: Run The Pipeline

### 6.1 Manual Trigger

1. Go to job: `merchant-backend-deploy`
2. Click **"Build Now"**
3. Watch the build progress

### 6.2 View Build Console

- Click on build number (e.g., #1)
- Click **"Console Output"**
- Watch real-time logs

### 6.3 Expected Stages

```
✅ Cleanup Workspace
✅ Clone Repository
✅ Verify Files
✅ Build Docker Image
✅ Stop Old Container
✅ Create Network
✅ Deploy Container
✅ Health Check
✅ Cleanup Old Images
```

---

## 🔍 Step 7: Verify Deployment

### 7.1 Check Container Status

```bash
# List containers
docker ps | grep merchant-backend

# Check logs
docker logs merchant-backend

# Check network
docker network inspect grafana-mysql-network
```

### 7.2 Test API

```bash
# Health check
curl http://localhost:8080/health

# Test API endpoint
curl -X POST http://localhost:8080/api/merchant/summary \
  -H "Content-Type: application/json" \
  -d '{"mid": ["000000000001"]}'
```

---

## 📊 Pipeline Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Repository                    │
│   https://github.com/aristit0/merchant_summary_...     │
└────────────────────┬────────────────────────────────────┘
                     │ git clone
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Jenkins Pipeline                       │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 1. Clone Repo                                    │  │
│  │ 2. Build Docker Image                            │  │
│  │ 3. Stop Old Container                            │  │
│  │ 4. Deploy New Container                          │  │
│  │ 5. Health Check                                  │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────┘
                     │ docker run
                     ▼
┌─────────────────────────────────────────────────────────┐
│              Docker Container (Port 8080)               │
│                 merchant-backend                        │
│            Network: grafana-mysql-network               │
└─────────────────────────────────────────────────────────┘
                     │
                     ▼ connects to
┌─────────────────────────────────────────────────────────┐
│                  Couchbase Server                       │
└─────────────────────────────────────────────────────────┘
```

---

## ⚙️ Environment Variables

Pipeline menggunakan environment variables ini:

```groovy
environment {
    GIT_REPO = 'https://github.com/aristit0/merchant_summary_demo_backend.git'
    GIT_BRANCH = 'main'
    
    DOCKER_IMAGE = 'merchant-summary-backend'
    DOCKER_CONTAINER = 'merchant-backend'
    DOCKER_NETWORK = 'grafana-mysql-network'
    DOCKER_PORT = '8080'
    
    COUCHBASE_HOST = 'host.docker.internal'
    COUCHBASE_USERNAME = 'admin'
    COUCHBASE_PASSWORD = 'T1ku$H1t4m'
}
```

**Customize sesuai kebutuhan Anda!**

---

## 🔧 Troubleshooting

### Issue 1: Docker permission denied

**Error:** `permission denied while trying to connect to Docker daemon`

**Fix:**
```bash
# Give Jenkins user docker permissions
docker exec -u root jenkins chmod 666 /var/run/docker.sock
```

### Issue 2: Cannot connect to Couchbase

**Error:** `failed to connect to Couchbase`

**Fix:**
```bash
# If Couchbase is on host machine
# Use: host.docker.internal

# If Couchbase is in Docker
# Use: couchbase container name or IP
docker network inspect grafana-mysql-network
```

### Issue 3: Port already in use

**Error:** `port 8080 already in use`

**Fix:**
```bash
# Stop old container
docker stop merchant-backend
docker rm merchant-backend

# Or use different port
# Change DOCKER_PORT in Jenkinsfile
```

### Issue 4: Build fails at Go build stage

**Error:** `go build failed`

**Fix:**
```bash
# Check Go version in Dockerfile
# Ensure go.mod and go.sum exist in repo
# Verify all Go files are committed
```

### Issue 5: Health check fails

**Error:** `Health check did not respond`

**Fix:**
```bash
# Check container logs
docker logs merchant-backend

# Check if Couchbase is accessible
docker exec merchant-backend ping host.docker.internal

# Verify environment variables
docker inspect merchant-backend | grep -A 10 "Env"
```

---

## 🔄 Auto-Deploy on Git Push

### Option 1: GitHub Webhook (Recommended)

**In GitHub:**
1. Go to repo → Settings → Webhooks
2. Add webhook
3. Payload URL: `http://YOUR_JENKINS_URL:8088/github-webhook/`
4. Content type: `application/json`
5. Events: `Just the push event`

**In Jenkins:**
1. Go to job configuration
2. Build Triggers: Check "GitHub hook trigger for GITScm polling"
3. Save

### Option 2: Poll SCM

Already configured in pipeline:
```groovy
triggers {
    pollSCM('H/5 * * * *')  // Check every 5 minutes
}
```

---

## 📋 Useful Commands

### Jenkins Management

```bash
# View Jenkins logs
docker logs -f jenkins

# Restart Jenkins
docker restart jenkins

# Backup Jenkins home
docker cp jenkins:/var/jenkins_home ./jenkins_backup
```

### Container Management

```bash
# View container logs
docker logs -f merchant-backend

# Execute command in container
docker exec -it merchant-backend sh

# Stop and remove
docker stop merchant-backend && docker rm merchant-backend

# View container details
docker inspect merchant-backend
```

### Network Management

```bash
# Inspect network
docker network inspect grafana-mysql-network

# List containers in network
docker network inspect grafana-mysql-network -f '{{range .Containers}}{{.Name}} {{end}}'

# Remove and recreate network
docker network rm grafana-mysql-network
docker network create grafana-mysql-network
```

---

## 📊 Monitoring & Logs

### View Pipeline Logs

```bash
# In Jenkins UI: Click build → Console Output
```

### View Container Logs

```bash
# Real-time logs
docker logs -f merchant-backend

# Last 100 lines
docker logs merchant-backend --tail 100

# Logs with timestamps
docker logs merchant-backend -t
```

### Check Container Health

```bash
# Container status
docker ps | grep merchant-backend

# Container stats
docker stats merchant-backend

# Container processes
docker top merchant-backend
```

---

## 🎯 Success Indicators

Pipeline is successful when:

1. ✅ All stages pass (green)
2. ✅ Container is running: `docker ps | grep merchant-backend`
3. ✅ Health check returns 200: `curl http://localhost:8080/health`
4. ✅ API responds: Test merchant summary endpoint
5. ✅ Logs show no errors: `docker logs merchant-backend`

---

## 🚀 Next Steps

After successful deployment:

1. **Test API**: Use Postman or curl
2. **Monitor logs**: Watch for any issues
3. **Setup alerts**: Configure Jenkins email notifications
4. **Add tests**: Add testing stage to pipeline
5. **Setup staging**: Create separate staging environment

---

## 📝 Pipeline Features

- ✅ **Automated Build**: Builds Docker image from source
- ✅ **Zero Downtime**: Stops old container, starts new one
- ✅ **Health Checks**: Verifies deployment success
- ✅ **Auto Cleanup**: Removes old Docker images
- ✅ **Network Management**: Ensures proper network configuration
- ✅ **Logging**: Comprehensive logs at each stage
- ✅ **Rollback Ready**: Keeps previous images for rollback

---

## 🎉 You're Done!

Your Jenkins CI/CD pipeline is now ready!

**Test it:**
```bash
# Make a change in your code
# Commit and push to GitHub
# Jenkins will auto-deploy!

# Or trigger manually in Jenkins UI
```

**Monitor:**
```bash
# Watch Jenkins build
# Check container logs
docker logs -f merchant-backend

# Test API
curl http://localhost:8080/health
```

**Enjoy your automated deployment! 🚀**