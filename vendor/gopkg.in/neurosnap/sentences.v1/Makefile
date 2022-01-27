BASE_DIR=$(shell echo $$GOPATH)/src/github.com/neurosnap/sentences
BINARY_DIR=./binary
CMD_DIR=./_cmd/sentences

VERSION_FILE=$(BASE_DIR)/VERSION
CURRENT_VERSION=$(shell cat $(VERSION_FILE))

COMMITHASH=$(shell git rev-parse --short HEAD)

.PHONY: english test build cross deploy install bump

test:
	go test ./...

build:
	go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}

get:
	go get ./...
	go get github.com/inconshreveable/mousetrap

cross:
	mkdir -p $(BINARY_DIR)

	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_linux-amd64.tar.gz ./sentences

	GOOS=linux GOARCH=386 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_linux-386.tar.gz ./sentences

	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_darwin-amd64.tar.gz ./sentences

	GOOS=darwin GOARCH=386 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_darwin-386.tar.gz ./sentences

	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_windows-amd64.tar.gz ./sentences

	GOOS=windows GOARCH=386 go build -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}
	tar -czvf $(BINARY_DIR)/sentences_windows-386.tar.gz ./sentences

deploy: cross
	source ~/virtualenvs/aws/bin/activate
	aws s3 cp ./binary s3://sentence-binaries/ --recursive --exclude "*" --include "*.tar.gz"

install:
	go install -ldflags "-X main.VERSION=$(CURRENT_VERSION) -X main.COMMITHASH=$(COMMITHASH)" ${CMD_DIR}

bump:
	MAJOR=$(word 1, $(subst ., , $(CURRENT_VERSION)))
	MINOR=$(word 2, $(subst ., , $(CURRENT_VERSION)))
	PATCH=$(word 3, $(subst ., , $(CURRENT_VERSION)))
	VER ?= $(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))

	echo $(VER) > $(VERSION_FILE)

# Compile language specific training data
czech:
	go-bindata -pkg="data" -o data/czech.go data/czech.json

danish:
	go-bindata -pkg="data" -o data/danish.go data/danish.json

dutch:
	go-bindata -pkg="data" -o data/dutch.go data/dutch.json

english:
	go-bindata -pkg="data" -o data/english.go data/english.json

estonian:
	go-bindata -pkg="data" -o data/estonian.go data/estonian.json

finnish:
	go-bindata -pkg="data" -o data/finnish.go data/finnish.json

french:
	go-bindata -pkg="data" -o data/french.go data/french.json

german:
	go-bindata -pkg="data" -o data/german.go data/german.json

greek:
	go-bindata -pkg="data" -o data/greek.go data/greek.json

italian:
	go-bindata -pkg="data" -o data/italian.go data/italian.json

norwegian:
	go-bindata -pkg="data" -o data/norwegian.go data/norwegian.json

polish:
	go-bindata -pkg="data" -o data/polish.go data/polish.json

portugues:
	go-bindata -pkg="data" -o data/potuguese.go data/portuguese.json

slovene:
	go-bindata -pkg="data" -o data/slovene.go data/slovene.json

spanish:
	go-bindata -pkg="data" -o data/spanish.go data/spanish.json

turkish:
	go-bindata -pkg="data" -o data/turkish.go data/turkish.json


