PKGS             := $(shell go list ./...)
test:
	@echo ">> TEST, \"full-mode\": race detector on"
	@$(foreach pkg, $(PKGS),\
		echo -n "     ";\
		go test -run '(Test|Example)' -race $(pkg) || exit 1;\
		)

