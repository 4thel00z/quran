ASSETS ?= ../quran-assets
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
SERVER ?= httpserver@hetzner
WEBROOT ?= private/quran.host

.PHONY: build test assets mirror install

build:
	go build -ldflags "-s -w -X github.com/4thel00z/quran/internal/cmd.version=$(VERSION)" -o bin/quran .

test:
	go test ./...

install:
	go install -ldflags "-s -w -X github.com/4thel00z/quran/internal/cmd.version=$(VERSION)" .

# Regenerate internal/assets/data from a TarteelAI/quran-assets checkout.
assets:
	go run ./tools/gen -assets $(ASSETS) -out internal/assets/data

# Copy every reciter's ayah MP3s into the quran.host web root on the server.
mirror: build
	./bin/quran mirror --all > bin/mirror.txt
	scp bin/mirror.txt deploy/mirror.sh $(SERVER):$(WEBROOT)/
	ssh $(SERVER) 'cd $(WEBROOT) && nohup ./mirror.sh www < mirror.txt > mirror.log 2>&1 &'
