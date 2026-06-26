.PHONY: install deps backend frontend run build deploy clean tidy

install:
	cd frontend && npm install

deps:
	go mod tidy
	go mod download
	cd frontend && npm install

tidy:
	go mod tidy

backend:
	ENV=dev go run .

frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

run: deps
	@echo "Starting Go backend and React frontend in dev mode..."
	@echo "Backend will run on http://localhost:8080"
	@echo "Frontend will run on http://localhost:3000"
	@make -j2 backend frontend

build: deps
	cd frontend && npm run build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

deploy:
	apps-platform app deploy --no-build

clean:
	cd frontend && rm -rf node_modules dist
	rm -f app
	go clean
