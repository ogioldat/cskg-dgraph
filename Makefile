GO ?= go
BIN_DIR := bin
CLIENT_PKG := ./cmd/client
CLIENT_BIN := $(BIN_DIR)/client

.PHONY: all client clean

all: client

client:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(CLIENT_BIN) $(CLIENT_PKG)

clean:
	@rm -f $(CLIENT_BIN)
