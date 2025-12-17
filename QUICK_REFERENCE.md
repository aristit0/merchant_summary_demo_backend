# ⚡ Jenkins CI/CD Quick Reference

## 🚀 Quick Start Commands

### Setup Jenkins (One Command)
```bash
./setup-jenkins.sh
```

### Or Docker Compose
```bash
docker-compose up -d
```

---

## 📋 Essential Commands

### Jenkins Management
```bash
# Start Jenkins
docker start jenkins

# Stop Jenkins
docker stop jenkins

# View logs
docker logs -f jenkins

# Get admin password
docker exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword

# Restart Jenkins
docker restart jenkins
```

### Backend Container
```bash
# View running containers
docker ps | grep merchant

# View logs
docker logs -f merchant-backend

# Stop container
docker stop merchant-backend

# Remove container
docker rm merchant-backend

# Rebuild and deploy
# Trigger Jenkins build or run manually:
docker build -t merchant-summary-backend:latest .
docker run -d --name merchant-backend \
  --network grafana-mysql-network \
  -p 8080:8080 \
  merchant-summary-backend:latest
```

### Network Management
```bash
# Create network
docker network create grafana-mysql-network

# Inspect network
docker network inspect grafana-mysql-network

# List containers in network
docker network inspect grafana-mysql-network \
  --format '{{range .Containers}}{{.Name}} {{end}}'
```

---

## 🌐 URLs

| Service | URL | Purpose |
|---------|-----|---------|
| Jenkins | http://localhost:8088 | CI/CD Dashboard |
| Backend API | http://localhost:8080 | Merchant API |
| Health Check | http://localhost:8080/health | API Health |

---

## 🔧 Jenkins Pipeline Stages

```
1. Cleanup Workspace      → Clean workspace
2. Clone Repository       → git clone from GitHub
3. Verify Files          → Check Go files exist
4. Build Docker Image    → docker build
5. Stop Old Container    → docker stop
6. Create Network        → docker network create
7. Deploy Container      → docker run
8. Health Check          → curl /health
9. Cleanup Old Images    → docker rmi old versions
```

---

## 🧪 Testing

### Test Health
```bash
curl http://localhost:8080/health
```

### Test API
```bash
curl -X POST http://localhost:8080/api/merchant/summary \
  -H "Content-Type: application/json" \
  -d '{"mid": ["000000000001"]}'
```

### Check Container Status
```bash
docker ps | grep merchant-backend
docker logs merchant-backend --tail 50
```

---

## 🔍 Troubleshooting

### Issue: Jenkins can't access Docker

**Fix:**
```bash
docker exec -u root jenkins chmod 666 /var/run/docker.sock
```

### Issue: Port 8080 already in use

**Fix:**
```bash
docker stop merchant-backend
docker rm merchant-backend
```

### Issue: Container can't connect to Couchbase

**Check:**
```bash
# Test from host
ping localhost

# Test from container
docker exec merchant-backend ping host.docker.internal
```

### Issue: Build fails

**Check:**
```bash
# View Jenkins logs
docker logs jenkins

# View container logs
docker logs merchant-backend

# Check network
docker network inspect grafana-mysql-network
```

---

## 📊 Monitoring

### Container Stats
```bash
docker stats merchant-backend
```

### View All Logs
```bash
docker logs merchant-backend -f --tail 100
```

### Check Resource Usage
```bash
docker system df
```

---

## 🔄 Deployment Workflow

```
1. Code Change → Push to GitHub
   ↓
2. Jenkins polls GitHub (every 5 min)
   ↓
3. Jenkins triggers build
   ↓
4. Build Docker image
   ↓
5. Stop old container
   ↓
6. Start new container
   ↓
7. Health check
   ↓
8. Done! ✅
```

---

## 🎯 Environment Variables

In Jenkinsfile:
```groovy
COUCHBASE_HOST = 'host.docker.internal'
COUCHBASE_USERNAME = 'admin'
COUCHBASE_PASSWORD = 'T1ku$H1t4m'
```

Customize these for your setup!

---

## 📝 Files Overview

```
jenkins-pipeline/
├── Jenkinsfile               ← Pipeline definition
├── Dockerfile                ← Go backend Dockerfile
├── docker-compose.yml        ← Complete stack
├── setup-jenkins.sh          ← Auto setup script
├── JENKINS_SETUP_GUIDE.md    ← Complete guide
└── QUICK_REFERENCE.md        ← This file
```

---

## 🚀 One-Line Deploy

```bash
# Setup everything
./setup-jenkins.sh && \
echo "Jenkins ready at http://localhost:8088"
```

---

## ✅ Success Checklist

- [ ] Jenkins running on :8088
- [ ] Network created
- [ ] Jenkins can access Docker
- [ ] Pipeline job created
- [ ] Build triggered successfully
- [ ] Container running on :8080
- [ ] Health check passes
- [ ] API responds correctly

---

## 📖 Full Documentation

See `JENKINS_SETUP_GUIDE.md` for:
- Complete setup instructions
- Detailed troubleshooting
- Advanced configuration
- Security best practices

---

**Everything you need to manage your Jenkins CI/CD! 🚀**