CMDS := $(notdir $(wildcard cmd/*))

.PHONY: build install clean $(CMDS)

build: $(CMDS)

$(CMDS):
	go build -o bin/$@ ./cmd/$@

install:
	@for cmd in $(CMDS); do \
		echo "Installing $$cmd..."; \
		go install ./cmd/$$cmd; \
	done

clean:
	rm -rf bin/
