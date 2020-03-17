all: install

install: go.sum
    GO111MODULE=on go install -tags "$(build_tags)" ./cmd/bidaoD
    GO111MODULE=on go install -tags "$(build_tags)" ./cmd/bidao

go.sum: go.mod
		@echo "dependencies unmodified"
    GO111MODULE=on @go mod verify
