PKGS             := $(shell go list ./...)

all: codequality test security build

test:
	@echo ">> TEST, \"full-mode\": race detector on"
	@$(foreach pkg, $(PKGS),\
		echo -n "     ";\
		go test -run '(Test|Example)' -race $(pkg) || exit 1;\
		)

bench:
	@echo ">> BENCHMARK"
	@go test -bench=. -benchmem ./...

codequality:
	@echo ">> CODE QUALITY"

	@echo -n "     GOLANGCI-LINTERS \n"
	@golangci-lint -v run ./...
	@$(call ok)

	@echo -n "     REVIVE"
	@revive -config revive.toml -formatter friendly -exclude vendor/... ./...
	@$(call ok)

security:
	@echo ">> CHECKING FOR INSECURE DEPENDENCIES USING GOVULNCHECK"
	@govulncheck ./...
	@echo ">> CHECKING FOR INSECURE DEPENDENCIES USING NANCY"
	@go list -json -deps | nancy sleuth
	@$(call ok)

build:
	@echo -n ">> BUILD"
	@npm install --prefix web
	@npm run build --prefix web
	@go build $(PKGS)
	@$(call ok)