_default:
  just --list

# Download module dependencies
[group: 'setup']
setup:
  go mod download

# Install the tmxu binary into $GOBIN (or $GOPATH/bin)
[group: 'install']
install:
  go install .

# Build the tmxu binary into the current directory
[group: 'build']
build:
  go build .

# Run tmxu with a command, e.g. just run version
[group: 'dev']
run CMD:
  go run ./ {{CMD}}

# Run lint, then tests; use before committing
[group: 'quality']
check: lint test

# Run all Go tests
[group: 'test']
test:
  go test -race ./...

# Run e2e checks on an isolated tmux server (needs tmux); optional prefix, e.g. just e2e 03
[group: 'test']
e2e PREFIX='':
  bash e2e/run.sh {{PREFIX}}

# Format Go code with gofmt
[group: 'quality']
fmt:
  go fmt ./...

# Format, then run go vet and go fix -diff
[group: 'quality']
lint: fmt
  go vet ./...
  go fix -diff ./...
