default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

# generate-mappings fetches the latest aztft map.json and regenerates
# internal/provider/azurerm_mappings.go. Requires network access.
generate-mappings:
	go run tools/gen_azurerm_mappings.go

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate generate-mappings
