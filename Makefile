.PHONY: test test-up test-down test-cover build vet fmt fmt-check

# --- Test environment ---------------------------------------------------------
TEST_ENV = \
	TEST_DB_HOST=127.0.0.1 \
	TEST_DB_PORT=3307 \
	TEST_DB_USER=root \
	TEST_DB_PASSWORD=testpassword \
	TEST_DB_NAME=expense_test

test-up:
	docker compose -f compose.test.yaml up -d --wait

test-down:
	docker compose -f compose.test.yaml down -v

test:
	$(TEST_ENV) go test ./... -count=1 -p 1

test-cover:
	$(TEST_ENV) go test ./... -count=1 -p 1 -coverprofile=coverage.out \
		-coverpkg=./repositories/...,./services/...,./controllers/...,./middlewares/...,./utilities/...,./dtos/...
	go tool cover -func=coverage.out | tail -1

# --- Static checks ------------------------------------------------------------
build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofmt -w $(shell gofmt -l . | grep -v '^docs/' | grep -v '^tmp/')

fmt-check:
	@test -z "$(shell gofmt -l . | grep -v '^docs/' | grep -v '^tmp/')" \
		|| (echo "unformatted files:" && gofmt -l . | grep -v '^docs/' | grep -v '^tmp/' && exit 1)
