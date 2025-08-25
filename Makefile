# go makefile

program != basename $$(pwd)

latest_release != gh 2>/dev/null release list --json tagName --jq '.[0].tagName' | tr -d v
version != cat VERSION

rstms_modules = $(shell awk <go.mod '/^module/{next} /rstms/{print $$1}')

gitclean = $(if $(shell git status --porcelain),$(error git status is dirty),$(info git status is clean))

$(program): build

build: fmt proxy_common_go
	fix go build . ./...
	go build

fmt: go.sum
	fix go fmt . ./...

go.mod:
	go mod init

go.sum: go.mod
	go mod tidy

install: build
	go install

test: fmt
	go test -v -failfast . ./...

debug: fmt
	go test -v -failfast -count=1 -run $(test) . ./...

release:
	$(gitclean)
	@$(if $(update),gh release delete -y v$(version),)
	gh release create v$(version) --notes "v$(version)"

update:
	@echo updating modules
	$(foreach module,$(rstms_modules),go get $(module)@latest;)

clean:
	rm -f $(program) *.core 
	go clean

sterile: clean
	which $(program) && go clean -i || true
	go clean
	go clean -cache
	go clean -modcache
	rm -f go.mod go.sum

proxy_common_go: *.go
	./generate_proxy_source >$@
