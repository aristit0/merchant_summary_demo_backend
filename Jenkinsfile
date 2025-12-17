pipeline {
    agent any

    options {
        skipDefaultCheckout(true)
        timestamps()
        disableConcurrentBuilds()
    }

    environment {
        // Git Configuration
        GIT_REPO   = 'https://github.com/aristit0/merchant_summary_demo_backend.git'
        GIT_BRANCH = 'main'

        // Docker Configuration
        DOCKER_IMAGE     = 'merchant-summary-backend'
        DOCKER_CONTAINER = 'merchant-backend'
        DOCKER_NETWORK   = 'grafana-mysql-network'
        DOCKER_PORT      = '8080'
        DOCKER_TAG       = "${BUILD_NUMBER}"

        // Couchbase Configuration
        COUCHBASE_HOST     = 'host.docker.internal'
        COUCHBASE_USERNAME = 'admin'
        COUCHBASE_PASSWORD = 'T1ku$H1t4m'
    }

    stages {

        stage('Prepare Workspace') {
            steps {
                echo '🧹 Preparing clean workspace...'
                deleteDir()
            }
        }

        stage('Checkout Source Code') {
            steps {
                echo '📥 Checking out source code...'
                git branch: "${GIT_BRANCH}", url: "${GIT_REPO}"
            }
        }

        stage('Verify Project Structure') {
            steps {
                echo '🔍 Verifying project files...'
                sh '''
                    echo "Workspace content:"
                    ls -la

                    echo "\nGo version:"
                    go version || echo "Go not required in Jenkins (Docker build only)"

                    echo "\nChecking go.mod:"
                    test -f go.mod && cat go.mod || (echo "❌ go.mod not found" && exit 1)
                '''
            }
        }

        stage('Build Docker Image') {
            steps {
                echo '🐳 Building Docker image...'
                sh '''
                    if [ ! -f Dockerfile ]; then
                        echo "Dockerfile not found, generating..."
                        cat > Dockerfile << 'EOF'
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o merchant-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/merchant-api .
EXPOSE 8080
CMD ["./merchant-api"]
EOF
                    fi

                    docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
                    docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
                '''
            }
        }

        stage('Stop Existing Container') {
            steps {
                echo '🛑 Stopping existing container (if any)...'
                sh '''
                    docker stop ${DOCKER_CONTAINER} 2>/dev/null || true
                    docker rm   ${DOCKER_CONTAINER} 2>/dev/null || true
                '''
            }
        }

        stage('Ensure Docker Network') {
            steps {
                echo '🌐 Ensuring Docker network exists...'
                sh '''
                    docker network inspect ${DOCKER_NETWORK} >/dev/null 2>&1 || \
                    docker network create ${DOCKER_NETWORK}
                '''
            }
        }

        stage('Deploy Application Container') {
            steps {
                echo '🚀 Deploying new container...'
                sh '''
                    docker run -d \
                        --name ${DOCKER_CONTAINER} \
                        --network ${DOCKER_NETWORK} \
                        -p ${DOCKER_PORT}:8080 \
                        -e COUCHBASE_HOST=${COUCHBASE_HOST} \
                        -e COUCHBASE_USERNAME=${COUCHBASE_USERNAME} \
                        -e COUCHBASE_PASSWORD=${COUCHBASE_PASSWORD} \
                        --restart unless-stopped \
                        ${DOCKER_IMAGE}:latest
                '''
            }
        }

        stage('Health Check') {
            steps {
                echo '🏥 Running health check...'
                sh '''
                    echo "Waiting service startup..."
                    sleep 5

                    docker ps | grep ${DOCKER_CONTAINER}

                    echo "Checking logs:"
                    docker logs ${DOCKER_CONTAINER} --tail 20

                    for i in {1..10}; do
                        if curl -sf http://localhost:${DOCKER_PORT}/health; then
                            echo "✅ Health check passed"
                            exit 0
                        fi
                        echo "Retry $i..."
                        sleep 3
                    done

                    echo "❌ Health check failed"
                    exit 1
                '''
            }
        }

        stage('Cleanup Old Images') {
            steps {
                echo '🧹 Cleaning old Docker images...'
                sh '''
                    docker images ${DOCKER_IMAGE} --format "{{.Tag}}" | \
                    grep -Ev "latest|<none>" | \
                    sort -rn | tail -n +4 | \
                    xargs -r docker rmi ${DOCKER_IMAGE}:{} || true
                '''
            }
        }
    }

    post {
        success {
            echo '✅ CI/CD Pipeline completed successfully!'
            echo "🌐 API Endpoint: http://localhost:${DOCKER_PORT}/health"
        }

        failure {
            echo '❌ CI/CD Pipeline failed'
            sh '''
                docker logs ${DOCKER_CONTAINER} --tail 50 || true
            '''
        }

        always {
            echo '📊 Final Docker State'
            sh '''
                echo "\nContainers:"
                docker ps -a | grep merchant || true

                echo "\nNetwork:"
                docker network inspect ${DOCKER_NETWORK} | grep -A 5 Containers || true
            '''
        }
    }
}