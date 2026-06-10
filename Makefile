all: build test

build:
	@echo "Building..."
	
	@go build -o main cmd/api/main.go

run:
	@go run cmd/api/main.go &
	@npm install --prefer-offline --no-fund --prefix ./frontend
	@npm run dev --prefix ./frontend

docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test:
	@echo "Testing..."
	@go test ./... -v


clean:
	@echo "Cleaning..."
	@rm -f main