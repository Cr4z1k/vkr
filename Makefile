OS     := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH   := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
PREFIX := /usr/local
BINDIR := $(PREFIX)/bin

URL := https://github.com/redpanda-data/redpanda/releases/latest/download/rpk-$(OS)-$(ARCH).zip

.PHONY: install-rpk

install-rpk:
	curl -sL $(URL) -o rpk.zip
	unzip -o rpk.zip rpk
	mv rpk $(BINDIR)/rpk
	chmod +x $(BINDIR)/rpk
	rm rpk.zip

brew-install-rpk:
	brew tap redpanda-data/tap
	brew install redpanda-data/tap/redpanda