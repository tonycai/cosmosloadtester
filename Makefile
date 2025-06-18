GO ?= "go"
LDFLAGS ?='-s -w'
GOOS ?= "linux"
GOARCH ?= "amd64"
CGO_ENABLED ?= 0

# Docker configuration
DOCKER_TAG ?= cosmosloadtester-cli:latest
DOCKER_REGISTRY ?= 

.PHONY: pb server ui cli docker docker-build docker-run docker-compose-up docker-compose-down

pb:
	(cd proto && buf mod update)
	(cd proto && buf generate --template buf.gen.yaml)
	(cd proto && buf generate --template buf.gen.ts.grpcweb.yaml --include-imports)
	npx --yes swagger-typescript-api -p ./proto/orijtech/cosmosloadtester/v1/loadtest_service.swagger.json -o ./ui/src/gen -n LoadtestApi.ts

server:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags $(LDFLAGS) -o bin/server ./cmd/server

cli:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags $(LDFLAGS) -o bin/cosmosloadtester-cli ./cmd/cli

ui:
	(cd ui && npm install)
	(cd ui && npm run build)

all: pb server cli ui

clean:
	rm -rf bin/
	rm -rf ui/build/

install-cli: cli
	cp bin/cosmosloadtester-cli /usr/local/bin/

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

# Development helpers
dev-deps:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker targets
docker-build:
	docker build -t $(DOCKER_TAG) .

docker-run: docker-build
	docker run -it --rm $(DOCKER_TAG)

docker-push: docker-build
	@if [ -n "$(DOCKER_REGISTRY)" ]; then \
		docker tag $(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_TAG); \
		docker push $(DOCKER_REGISTRY)/$(DOCKER_TAG); \
	else \
		echo "DOCKER_REGISTRY not set, skipping push"; \
	fi

# Docker Compose targets
docker-compose-up:
	docker-compose up -d cosmosloadtester-cli

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f cosmosloadtester-cli

docker-compose-exec:
	docker-compose exec cosmosloadtester-cli /bin/sh

docker-compose-build:
	docker-compose build cosmosloadtester-cli

docker-compose-rebuild:
	docker-compose build --no-cache cosmosloadtester-cli

# Development with Docker
docker-dev: docker-compose-build docker-compose-up

# Full monitoring stack
docker-monitoring:
	docker-compose --profile monitoring up -d

# Docker cleanup
docker-clean:
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

# Create directories for Docker volumes
docker-init:
	mkdir -p config results testnet-data monitoring

# Docker testing
docker-test: docker-compose-up
	@echo "Running basic Docker test..."
	docker-compose exec -T cosmosloadtester-cli cosmosloadtester-cli --version
	docker-compose exec -T cosmosloadtester-cli cosmosloadtester-cli --list-factories
	@echo "Docker test completed successfully!"

.PHONY: all clean install-cli test fmt lint dev-deps docker-build docker-run docker-push docker-compose-up docker-compose-down docker-compose-logs docker-compose-exec docker-compose-build docker-compose-rebuild docker-dev docker-monitoring docker-clean docker-init docker-test
