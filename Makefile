version=$(shell git describe --tags)
buildinfo_pkg=gitlab.joelpet.se/joelpet/go-rebalance/internal/buildinfo

LDFLAGS=-X '$(buildinfo_pkg).Version=$(version)'
build_flags=$(if $(LDFLAGS),-ldflags="$(LDFLAGS)")

.PHONY: all
all: build

out: ; mkdir -p $@

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build: | out
	go build -o out/ $(build_flags) ./cmd/rebalance

.PHONY: install
install:
		go install $(build_flags) ./cmd/rebalance

.PHONY: clean
clean:
	-rm -r out
