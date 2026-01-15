_default:
  just --list

# Run app
[group: 'cli']
run CMD:
  go run ./cli CMD

# Run all tests
[group: 'tests']
test:
	go test ./... -v

