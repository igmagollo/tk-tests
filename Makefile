.PHONY: run test e2e docker tk-workflows tk-secret tk-apply tk-run

# Namespace the Testkube agent runs executions in.
TK_NAMESPACE ?= tk-agent
# Name/key of the secret holding the GitHub token the workflows clone with.
TK_GIT_SECRET ?= tk-tests-git
# Runner label selector: matches the in-cluster agent, which is the only runner
# that can read the token secret above.
TK_TARGET ?= testkube.io/source=cloud

run:
	go run ./cmd/server

test:
	go test ./... -v

e2e:
	./test/e2e/api_test.sh

docker:
	docker build -t tk-tests:local .

# Regenerate testkube/*.yaml. Pass MODE=files to inline the sources instead of
# cloning the repo.
tk-workflows:
	python3 scripts/gen-workflows.py $(if $(MODE),--mode $(MODE),)

# Create/update the GitHub token secret the workflows clone with:
#   make tk-secret GITHUB_TOKEN=ghp_xxx
tk-secret:
	@test -n "$(GITHUB_TOKEN)" || { echo "set GITHUB_TOKEN=<PAT>"; exit 1; }
	kubectl -n $(TK_NAMESPACE) create secret generic $(TK_GIT_SECRET) \
		--from-literal=token=$(GITHUB_TOKEN) \
		--dry-run=client -o yaml | kubectl apply -f -

tk-apply: tk-workflows
	kubectl testkube create testworkflow --update -f testkube/go-unit-tests.yaml
	kubectl testkube create testworkflow --update -f testkube/go-api-e2e.yaml

tk-run: tk-apply
	kubectl testkube run testworkflow go-unit-tests --target $(TK_TARGET) -f
	kubectl testkube run testworkflow go-api-e2e --target $(TK_TARGET) -f
