.PHONY: run test e2e docker tk-workflows tk-apply tk-run

run:
	go run ./cmd/server

test:
	go test ./... -v

e2e:
	./test/e2e/api_test.sh

docker:
	docker build -t tk-tests:local .

# Regenerate testkube/*.yaml from the sources in this repo.
tk-workflows:
	python3 scripts/gen-workflows.py

tk-apply: tk-workflows
	kubectl testkube create testworkflow --update -f testkube/go-unit-tests.yaml
	kubectl testkube create testworkflow --update -f testkube/go-api-e2e.yaml

tk-run: tk-apply
	kubectl testkube run testworkflow go-unit-tests -f
	kubectl testkube run testworkflow go-api-e2e -f
