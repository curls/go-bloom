all: deps test
	@echo "Building go-bloom..."
	@go fmt
	@go build

test: deps
	@echo "Running tests..."
	@go fmt
	@go test -run .
	@go test -bench .

deps:
	@echo "Fetching dependencies..."
