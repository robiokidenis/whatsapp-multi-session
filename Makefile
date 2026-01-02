.PHONY: help init setup deploy start stop restart logs status clean fix-permissions build kill docker-build docker-buildx-init docker-buildx docker-push docker-publish docker-publish-multi docker-login deploy-image deploy-image-start

# Docker Configuration
DOCKER_REGISTRY ?= docker.io
DOCKER_IMAGE ?= rod16/whatsapp-multi-session
DOCKER_TAG ?= latest

# Default target
help: ## Show this help message
	@echo "WhatsApp Multi-Session Manager - Deployment Commands"
	@echo "=================================================="
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

init: ## Initialize the project (run once)
	@echo "🚀 Initializing WhatsApp Multi-Session Manager..."
	@chmod +x docker-init.sh
	@./docker-init.sh

setup: init ## Alias for init

fix-permissions: ## Fix directory permissions for Docker container
	@echo "🔧 Fixing directory permissions..."
	@if [ -f ./fix-docker-permissions.sh ]; then \
		sudo ./fix-docker-permissions.sh; \
	else \
		sudo chown -R 1001:1001 ./whatsapp ./config || echo "⚠️  Failed to set permissions. Run manually: sudo chown -R 1001:1001 ./whatsapp ./config"; \
		sudo chmod -R u+w ./whatsapp; \
		echo "✅ Permissions fixed"; \
	fi

build: ## Build Docker images
	@echo "🏗️  Building Docker images..."
	@docker-compose build

deploy: fix-permissions build ## Deploy the application (fix permissions, build, and start)
	@echo "🚀 Deploying WhatsApp Multi-Session Manager..."
	@docker-compose up -d
	@echo "✅ Deployment complete!"
	@$(MAKE) status

deploy-prod: ## Deploy for production (handles Ubuntu/Linux permission issues)
	@echo "🏭 Deploying for production environment..."
	@chmod +x deploy-production.sh
	@./deploy-production.sh

start: fix-permissions ## Start the application
	@echo "▶️  Starting WhatsApp Multi-Session Manager..."
	@docker-compose up -d
	@$(MAKE) status

stop: ## Stop the application
	@echo "⏹️  Stopping WhatsApp Multi-Session Manager..."
	@docker-compose down

restart: ## Restart the application
	@echo "🔄 Restarting WhatsApp Multi-Session Manager..."
	@docker-compose down
	@$(MAKE) fix-permissions
	@docker-compose up -d
	@$(MAKE) status

logs: ## Show application logs
	@echo "📋 Showing logs (press Ctrl+C to exit)..."
	@docker-compose logs -f

status: ## Show application status
	@echo "📊 Application Status:"
	@echo "===================="
	@docker-compose ps
	@echo ""
	@echo "🌐 Application URL: http://localhost:$$(grep -E '^PORT=' .env 2>/dev/null | cut -d'=' -f2 || echo '8080')"
	@echo "👤 Default login: admin / admin123"

clean: ## Clean up containers and images
	@echo "🧹 Cleaning up..."
	@docker-compose down --volumes --remove-orphans
	@docker system prune -f

# Development commands
dev-logs: ## Show development logs (last 50 lines)
	@docker-compose logs --tail=50

dev-shell: ## Open shell in container
	@docker-compose exec whatsapp-multi-session /bin/sh

# Backup and restore
backup: ## Create backup of data
	@echo "💾 Creating backup..."
	@mkdir -p backups
	@tar -czf backups/whatsapp-backup-$$(date +%Y%m%d-%H%M%S).tar.gz whatsapp config .env
	@echo "✅ Backup created in backups/ directory"

# Health check
health: ## Check application health
	@echo "🏥 Checking application health..."
	@curl -s http://localhost:$$(grep -E '^PORT=' .env 2>/dev/null | cut -d'=' -f2 || echo '8080')/api/health || echo "❌ Health check failed"

# Update
update: ## Update and redeploy
	@echo "🔄 Updating application..."
	@git pull
	@$(MAKE) deploy

# Quick commands
up: start ## Alias for start
down: stop ## Alias for stop

# Development commands for local run
run: ## Run the application locally (loads .env)
	@echo "🚀 Running application locally..."
	@go run main.go

run-watch: ## Run with auto-reload (requires air)
	@echo "🚀 Running with auto-reload..."
	@air

kill:
	@PID=$$(lsof -ti tcp:8080); \
	if [ -n "$$PID" ]; then \
		echo "Killing process on port 8080 (PID: $$PID)"; \
		kill -9 $$PID; \
	else \
		echo "No process found on port 8080"; \
	fi

# Docker Hub commands
docker-login: ## Login to Docker Hub
	@echo "🔐 Logging in to Docker Hub..."
	@docker login

docker-build: ## Build Docker image for Docker Hub (single arch)
	@echo "🏗️  Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Image built successfully: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker images | grep $(DOCKER_IMAGE)

docker-buildx-init: ## Initialize Docker buildx for multi-arch builds
	@echo "🔧 Initializing Docker buildx..."
	@docker buildx create --name multiarch --use 2>/dev/null || true
	@docker buildx inspect --bootstrap

docker-buildx: ## Build multi-architecture Docker image (amd64, arm64)
	@echo "🏗️  Building multi-arch Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "📦 Platforms: linux/amd64, linux/arm64"
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		--push \
		.
	@echo "✅ Multi-arch image built and pushed: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "🌐 Image URL: https://hub.docker.com/r/$(DOCKER_IMAGE)"

docker-push: ## Push Docker image to Docker Hub
	@echo "📤 Pushing image to Docker Hub: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Image pushed successfully!"
	@echo "🌐 Image URL: https://hub.docker.com/r/$(DOCKER_IMAGE)"

docker-publish: docker-build docker-push ## Build and push Docker image (full publish workflow)
	@echo "🎉 Docker image published: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-publish-multi: docker-buildx ## Publish multi-architecture image to Docker Hub
	@echo "🎉 Multi-arch Docker image published: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-tag-latest: docker-build ## Tag current build as latest
	@echo "🏷️  Tagging as latest..."
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "✅ Tagged as $(DOCKER_IMAGE):latest"

docker-push-all: docker-build ## Build and push all tags (latest and version tag)
	@echo "📤 Pushing all tags..."
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest
	@echo "✅ All tags pushed successfully!"

# Docker Image-based deployment (uses pre-built image from Docker Hub)
deploy-image: ## Deploy using pre-built Docker image from Docker Hub
	@echo "🚀 Deploying WhatsApp Multi-Session Manager from Docker Hub image..."
	@echo "📦 Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@if [ ! -f .env ]; then \
		echo "⚠️  Warning: .env file not found. Creating default .env..."; \
		cp .env.example .env 2>/dev/null || echo "PORT=8080" > .env; \
	fi
	@mkdir -p whatsapp config logs
	@docker-compose -f docker-compose.image.yml pull
	@docker-compose -f docker-compose.image.yml up -d
	@echo "✅ Deployment complete!"
	@$(MAKE) status-image

deploy-image-start: ## Start with Docker Hub image (without pulling)
	@echo "▶️  Starting WhatsApp Multi-Session Manager from Docker Hub image..."
	@docker-compose -f docker-compose.image.yml up -d
	@$(MAKE) status-image

stop-image: ## Stop Docker Hub image deployment
	@echo "⏹️  Stopping WhatsApp Multi-Session Manager..."
	@docker-compose -f docker-compose.image.yml down

logs-image: ## Show logs from Docker Hub image deployment
	@echo "📋 Showing logs (press Ctrl+C to exit)..."
	@docker-compose -f docker-compose.image.yml logs -f

restart-image: stop-image deploy-image-start ## Restart Docker Hub image deployment

status-image: ## Show status of Docker Hub image deployment
	@echo "📊 Application Status:"
	@echo "===================="
	@docker-compose -f docker-compose.image.yml ps
	@echo ""
	@echo "🌐 Application URL: http://localhost:$$(grep -E '^PORT=' .env 2>/dev/null | cut -d'=' -f2 || echo '8080')"
	@echo "👤 Default login: admin / admin123"

update-image: ## Pull latest image and restart
	@echo "🔄 Updating to latest image..."
	@docker-compose -f docker-compose.image.yml pull
	@docker-compose -f docker-compose.image.yml up -d
	@echo "✅ Update complete!"
