BINARY_NAME=openPipe
COVERAGE_FILE=coverage.out

.PHONY: test coverage coverage.html clean build

test:
	@echo "Running tests with coverage..."
	@go test -covermode=count -coverprofile=$(COVERAGE_FILE) ./...
	@# Filter out main.go lines
	@grep -v "main.go:" $(COVERAGE_FILE) > $(COVERAGE_FILE).filtered
	@mv $(COVERAGE_FILE).filtered $(COVERAGE_FILE)
	@echo "Tests complete."

coverage:
	@echo "Coverage summary (excluding main.go):"
	@go tool cover -func=$(COVERAGE_FILE) | grep total:

coverage.html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated at coverage.html"

clean:
	@echo "Cleaning coverage files and binary..."
	@rm -f $(COVERAGE_FILE) coverage.html $(BINARY_NAME)
	@echo "Clean complete."

build:
	@echo "Building binary $(BINARY_NAME)..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME) .
	@echo "Build complete."

