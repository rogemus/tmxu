_default:
  just --list

# Run tmxu with a command, e.g. just run version
[group: 'dev']
run CMD:
  go run ./ {{CMD}}

# Run all Go tests
[group: 'test']
test:
  go test ./...

# Format Go code with gofmt
[group: 'quality']
fmt:
  go fmt ./...

# Format, then run go vet and go fix -diff
[group: 'quality']
lint: fmt
  go vet ./...
  go fix -diff ./...
