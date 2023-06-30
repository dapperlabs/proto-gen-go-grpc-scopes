.PHONY: generate-proto
generate-proto:
	@echo "Generating proto files..."
	@protoc --proto_path=proto --go_out=scopesproto --go_opt=paths=source_relative proto/scopes.proto

.PHONY: generate-test-proto
generate-test-proto:
	@echo "Generating test proto files..."
	@cd test && buf mod update && buf generate

.PHONY: test
test: generate-test-proto
	@echo "Running tests..."
	@go test ./...