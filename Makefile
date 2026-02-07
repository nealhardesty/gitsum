BINARY := gitsum
VERSION := $(shell grep 'const Version' version.go | cut -d'"' -f2)

.PHONY: build test run clean lint fmt tidy version version-increment release help

## build: Compile the project
build:
	go build -o $(BINARY) .

## test: Run all tests with race detection
test:
	go test -race -v ./...

## run: Build and run the application
run: build
	./$(BINARY)

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	go clean

## lint: Run linters (go vet)
lint:
	go vet ./...

## fmt: Format code with gofmt
fmt:
	gofmt -w .

## tidy: Run go mod tidy
tidy:
	go mod tidy

## version: Display current version
version:
	@echo "$(BINARY) v$(VERSION)"

## version-increment: Increment patch version
version-increment:
	@echo "Current version: $(VERSION)"; \
	MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	MINOR=$$(echo $(VERSION) | cut -d. -f2); \
	PATCH=$$(echo $(VERSION) | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	sed -i "s/const Version = \"$(VERSION)\"/const Version = \"$$NEW_VERSION\"/" version.go; \
	echo "Version updated to $$NEW_VERSION"

## release: Create a git tag for the current version
release: build
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Tagged v$(VERSION). Run 'git push origin v$(VERSION)' to publish."

## help: Show this help message
help:
	@echo "$(BINARY) v$(VERSION) - Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
