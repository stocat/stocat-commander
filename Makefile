.PHONY: all run check-go

# Default target
all: run

# Run the Stocat Commander TUI
run: check-go
	@echo "Starting Stocat Commander..."
	@go run main.go

# Check if Go is installed, and install via Homebrew if missing
check-go:
	@if ! command -v go > /dev/null 2>&1; then \
		echo "Go is not installed. Attempting to install via Homebrew..."; \
		if command -v brew > /dev/null 2>&1; then \
			brew install go; \
		else \
			echo "Error: Homebrew is not installed. Please install Go manually."; \
			exit 1; \
		fi; \
	fi
