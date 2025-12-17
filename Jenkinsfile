pipeline {
    agent any
    
    environment {
        // GitHub Repository
        GIT_REPO = 'https://github.com/aristit0/merchant_summary_demo_backend.git'
        GIT_BRANCH = 'main'
        
        // Docker Configuration
        DOCKER_IMAGE = 'merchant-summary-backend'
        DOCKER_TAG = "${BUILD_NUMBER}"
        DOCKER_CONTAINER = 'merchant-backend'
        DOCKER_NETWORK = 'grafana-mysql-network'
        DOCKER_PORT = '8080'
        
        // Couchbase Configuration (adjust these)
        COUCHBASE_HOST = 'host.docker.internal'
        COUCHBASE_USERNAME = 'admin'
        COUCHBASE_PASSWORD = 'T1ku$H1t4m'
    }
    
    stages {
        stage('Cleanup Workspace') {
            steps {
                echo '🧹 Cleaning up workspace...'
                deleteDir()
            }
        }
        
        stage('Clone Repository') {
            steps {
                echo '📥 Cloning repository from GitHub...'
                git branch: "${GIT_BRANCH}", url: "${GIT_REPO}"
            }
        }
        
        stage('Verify Files') {
            steps {
                echo '🔍 Verifying project files...'
                sh '''
                    echo "Files in workspace:"
                    ls -la
                    
                    echo "\nChecking Go files:"
                    ls -la *.go || echo "No Go files in root"
                    
                    echo "\nChecking go.mod:"
                    cat go.mod || echo "No go.mod found"
                '''
            }
        }
        
        stage('Build Docker Image') {
            steps {
                echo '🐳 Building Docker image...'
                script {
                    // Create Dockerfile if not exists in repo
                    sh '''
                        if [ ! -f Dockerfile ]; then
                            echo "Creating Dockerfile..."
                            cat > Dockerfile << 'EOF'
# Multi-stage build for Go backend
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o merchant-api .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/merchant-api .

EXPOSE 8080

CMD ["./merchant-api"]
EOF
                        fi
                    '''
                    
                    // Build Docker image
                    sh """
                        docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
                        docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
                    """
                }
            }
        }
        
        stage('Stop Old Container') {
            steps {
                echo '🛑 Stopping and removing old container...'
                script {
                    sh """
                        # Stop container if exists
                        docker stop ${DOCKER_CONTAINER} || true
                        
                        # Remove container if exists
                        docker rm ${DOCKER_CONTAINER} || true
                        
                        echo "Old container removed"
                    """
                }
            }
        }
        
        stage('Create Network') {
            steps {
                echo '🌐 Ensuring Docker network exists...'
                script {
                    sh """
                        # Create network if not exists
                        docker network inspect ${DOCKER_NETWORK} >/dev/null 2>&1 || \
                        docker network create ${DOCKER_NETWORK}
                        
                        echo "Network ${DOCKER_NETWORK} is ready"
                    """
                }
            }
        }
        
        stage('Deploy Container') {
            steps {
                echo '🚀 Deploying new container...'
                script {
                    sh """
                        docker run -d \
                            --name ${DOCKER_CONTAINER} \
                            --network ${DOCKER_NETWORK} \
                            -p ${DOCKER_PORT}:8080 \
                            -e COUCHBASE_HOST=${COUCHBASE_HOST} \
                            -e COUCHBASE_USERNAME=${COUCHBASE_USERNAME} \
                            -e COUCHBASE_PASSWORD=${COUCHBASE_PASSWORD} \
                            --restart unless-stopped \
                            ${DOCKER_IMAGE}:latest
                        
                        echo "Container deployed successfully"
                    """
                }
            }
        }
        
        stage('Health Check') {
            steps {
                echo '🏥 Performing health check...'
                script {
                    sh '''
                        echo "Waiting for container to start..."
                        sleep 5
                        
                        echo "Checking container status:"
                        docker ps | grep merchant-backend
                        
                        echo "\nChecking container logs:"
                        docker logs merchant-backend --tail 20
                        
                        echo "\nTesting health endpoint:"
                        for i in {1..10}; do
                            if curl -f http://localhost:8080/health; then
                                echo "\n✅ Health check passed!"
                                exit 0
                            fi
                            echo "\nRetrying in 3 seconds... ($i/10)"
                            sleep 3
                        done
                        
                        echo "\n⚠️ Health check did not respond, but container is running"
                        echo "Check logs: docker logs merchant-backend"
                    '''
                }
            }
        }
        
        stage('Cleanup Old Images') {
            steps {
                echo '🧹 Cleaning up old Docker images...'
                script {
                    sh """
                        # Keep only the last 3 builds
                        docker images ${DOCKER_IMAGE} --format "{{.Tag}}" | \
                        grep -v latest | \
                        sort -rn | \
                        tail -n +4 | \
                        xargs -I {} docker rmi ${DOCKER_IMAGE}:{} || true
                        
                        echo "Cleanup completed"
                    """
                }
            }
        }
    }
    
    post {
        success {
            echo '✅ Pipeline completed successfully!'
            echo "🌐 Backend API available at: http://localhost:${DOCKER_PORT}"
            echo "🔍 Test with: curl http://localhost:${DOCKER_PORT}/health"
        }
        failure {
            echo '❌ Pipeline failed!'
            echo '📋 Check logs: docker logs ${DOCKER_CONTAINER}'
        }
        always {
            echo '📊 Pipeline execution completed'
            sh '''
                echo "\n=== Container Status ==="
                docker ps -a | grep merchant-backend || echo "No container found"
                
                echo "\n=== Network Status ==="
                docker network inspect grafana-mysql-network | grep -A 5 "Containers" || echo "Network not found"
            '''
        }
    }
}