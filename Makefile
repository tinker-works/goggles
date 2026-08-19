GO := go

.PHONY: fmt fmt-check mod test vet lint build

fmt:
	$(GO)fmt -w $$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$$.Dir}}/{{.}} {{end}}{{range .XTestGoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...)

fmt-check:
	@test -z "$$($(GO)fmt -l $$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$$.Dir}}/{{.}} {{end}}{{range .XTestGoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...))" || { echo "Go files need formatting"; exit 1; }

mod:
	$(GO) mod verify

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

build:
	$(GO) build ./...
