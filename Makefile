
test:
	@go test -count=1 ./...

test-coverage:
	@go test -coverprofile=cover.out -count=1 ./...
	@go tool cover -html=cover.out
