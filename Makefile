GOBUILD ?= go build

all: cmd/opensend/main.go
	$(GOBUILD) ./cmd/opensend

install: opensend opensend.toml
	install -Dm755 opensend $(DESTDIR)/usr/bin/opensend
	install -Dm644 opensend.toml $(DESTDIR)/etc/opensend.toml

install-macos: opensend opensend.toml
	mkdir -p $(DESTDIR)/usr/local/bin
	install -m755 opensend $(DESTDIR)/usr/local/bin/opensend
	mkdir -p $(DESTDIR)/etc
	install -m644 opensend.toml $(DESTDIR)/etc/opensend.toml