# Run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
templ:
	go tool templ generate --watch --proxy="http://localhost:8090" --open-browser=false

# Run air to detect any go file changes to re-build and re-run the server.
server:
	go tool air \
	--build.cmd "go build -o tmp/bin/main ./cmd/server/main.go" \
	--build.bin "tmp/bin/main" \
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

tailwind-clean:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --clean

# Run tailwindcss to generate the styles.css bundle in watch mode.
tailwind-watch:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

# Start development server
dev:
	@echo "Starting services (MinIO, PostgreSQL if enabled)..."
	@docker compose up -d
	@sleep 2
	@echo "Starting app..."
	@make tailwind-clean
	@make -j3 tailwind-watch templ server

# Stop all services
down:
	docker compose down

