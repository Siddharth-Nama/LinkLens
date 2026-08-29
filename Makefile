.PHONY: run test vet fmt docker-build

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

docker-build:
	docker build -t linklens .
