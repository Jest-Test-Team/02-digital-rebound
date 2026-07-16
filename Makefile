.PHONY: run test test-unit test-robot build tidy clean

ADDR ?= :8081
SCHEMA_PATH ?= schemas/event.schema.json
EVIDENCE_DIR ?= data/evidence
ASSESSMENT_DIR ?= data/assessments
API_BASE_URL ?= http://127.0.0.1:8081

run:
	cd src && ADDR=$(ADDR) SCHEMA_PATH=../$(SCHEMA_PATH) EVIDENCE_DIR=../$(EVIDENCE_DIR) ASSESSMENT_DIR=../$(ASSESSMENT_DIR) go run ./cmd/server

build:
	cd src && go build -o ../bin/digital-rebound-api ./cmd/server

tidy:
	cd src && go mod tidy

test-unit:
	cd src && go test ./...

test-robot:
	./scripts/run-robot-tests.sh

test: test-unit test-robot

clean:
	rm -rf bin data/evidence/* data/assessments/* tests/robot/results
