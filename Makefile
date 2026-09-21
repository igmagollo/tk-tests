.PHONY: run test e2e docker

run:
	go run ./cmd/server

test:
	go test ./... -v

e2e:
	./test/e2e/api_test.sh

docker:
	docker build -t tk-tests:local .
