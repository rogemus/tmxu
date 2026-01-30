_default:
  just --list

# Run app
[group: 'cli']
run CMD:
  go run ./ CMD

# Run all tests
[group: 'tests']
test:
	go test ./... -v

# Format code
fmt:
  go fmt ./... 

# Lint
lint: fmt 
  go vet ./...
