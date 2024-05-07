OS = $(shell uname | tr '[:upper:]' '[:lower:]')
SHELL:=/bin/bash

WORK_DIR=$(shell pwd)
BUILD_DIR=$(WORK_DIR)/build
PKG_CONFIG_DIR=$(BUILD_DIR)/pkgconfig

LIBOQS_DIR=$(BUILD_DIR)/liboqs
LIBOQS_OBJ=$(LIBOQS_DIR)/lib/liboqs.a
LIBOQS_PKG_CONFIG=$(PKG_CONFIG_DIR)/liboqs.pc

LIBCRYPTO_OBJ=$(LIBOQS_DIR)/lib/libcrypto.a
LIBSSL_OBJ=$(LIBOQS_DIR)/lib/libssl.a

GO_DEMO_BIN=$(BUILD_DIR)/abelsdk-demo

build: $(GO_DEMO_BIN)

$(GO_DEMO_BIN): $(LIBOQS_PKG_CONFIG)
	@echo "==> Building Abelian Go SDK demo ..."
	PKG_CONFIG_PATH=$(PKG_CONFIG_DIR) go build -o $(GO_DEMO_BIN) --tags=demo

clean:
	rm -rf $(GO_DEMO_BIN)

clean-all:
	rm -rf $(GO_DEMO_BIN) $(PKG_CONFIG_DIR) $(LIBCRYPTO_OBJ) $(LIBSSL_OBJ)

ifeq ($(OS), darwin)

$(LIBOQS_PKG_CONFIG): $(LIBOQS_OBJ) $(LIBCRYPTO_OBJ) $(LIBSSL_OBJ)
	@echo "==> Generating liboqs.pc ..."
	@if [ ! -d "${PKG_CONFIG_DIR}" ]; then mkdir -p ${PKG_CONFIG_DIR}; fi
	@echo "Name: liboqs" > $(LIBOQS_PKG_CONFIG)
	@echo "Description: C library for quantum resistant cryptography" >> $(LIBOQS_PKG_CONFIG)
	@echo "Version: 0.7.2-dev" >> $(LIBOQS_PKG_CONFIG)
	@echo "Cflags: -I$(LIBOQS_DIR)/include" >> $(LIBOQS_PKG_CONFIG)
	@echo "Ldflags: '-extldflags \"-static -Wl,-stack_size -Wl,0x1000000\"'" >> $(LIBOQS_PKG_CONFIG)
	@echo "Libs: -L$(LIBOQS_DIR)/lib -loqs -lcrypto -lssl" >> $(LIBOQS_PKG_CONFIG)

OPENSSL_SEARCH_PATHS = \
	/opt/homebrew/opt/openssl@1.1 \
	/opt/homebrew/opt/openssl@3 \
	/usr/local/opt/openssl@1.1 \
	/usr/local/opt/openssl@3 \
	/usr/local/Cellar/openssl@1.1 \
	/usr/local/Cellar/openssl@3

$(LIBCRYPTO_OBJ):
	@echo "==> Searching libcrypto.a ..."
	@for path in $(OPENSSL_SEARCH_PATHS); do \
		if [ -f "$$path/lib/libcrypto.a" ]; then \
			echo "Found libcrypto.a at $$path/lib/libcrypto.a."; \
			cp $$path/lib/libcrypto.a $(LIBCRYPTO_OBJ); \
			break; \
		fi; \
	done
	@if [ ! -f "$(LIBCRYPTO_OBJ)" ]; then \
		echo "*** Could not find libcrypto.a in your system. Please install OpenSSL or manually copy the file to $(LIBCRYPTO_OBJ)."; \
		exit 1; \
	fi

$(LIBSSL_OBJ):
	@echo "==> Searching libssl.a ..."
	@for path in $(OPENSSL_SEARCH_PATHS); do \
		if [ -f "$$path/lib/libssl.a" ]; then \
			echo "Found libssl.a at $$path/lib/libssl.a."; \
			cp $$path/lib/libssl.a $(LIBSSL_OBJ); \
			break; \
		fi; \
	done
	@if [ ! -f "$(LIBSSL_OBJ)" ]; then \
		echo "*** Could not find libssl.a in your system. Please install OpenSSL or manually copy the file to $(LIBSSL_OBJ)."; \
		exit 1; \
	fi

else

$(LIBOQS_PKG_CONFIG): $(LIBOQS_OBJ)
	@echo "==> Generating liboqs.pc ..."
	@if [ ! -d "${PKG_CONFIG_DIR}" ]; then mkdir -p ${PKG_CONFIG_DIR}; fi
	@echo "Name: liboqs" > $(LIBOQS_PKG_CONFIG)
	@echo "Description: C library for quantum resistant cryptography" >> $(LIBOQS_PKG_CONFIG)
	@echo "Version: 0.7.2-dev" >> $(LIBOQS_PKG_CONFIG)
	@echo "Cflags: -I$(LIBOQS_DIR)/include" >> $(LIBOQS_PKG_CONFIG)
	@echo "Ldflags: '-extldflags \"-static -Wl,-stack_size -Wl,0x1000000\"'" >> $(LIBOQS_PKG_CONFIG)
	@echo "Libs: -L$(LIBOQS_DIR)/lib -l:liboqs.a -lcrypto" >> $(LIBOQS_PKG_CONFIG)

endif

$(LIBOQS_OBJ):
	@if [ ! -d "${BUILD_DIR}" ]; then mkdir -p ${BUILD_DIR}; fi
	@if [ ! -d "${LIBOQS_DIR}" ]; then echo "==> Fetching liboqs ..."; git clone https://github.com/cryptosuite/liboqs.git ${LIBOQS_DIR}; fi
	@echo "==> Compiling liboqs ..."
	cd ${LIBOQS_DIR} && cmake -GNinja . && ninja

setenv:
	go env -w GOPRIVATE=github.com/pqabelian/*
	git config --global url."git@github.com:".insteadOf https://github.com/

unsetenv:
	go env -w GOPRIVATE=
	git config --global --unset url."git@github.com:".insteadOf https://github.com/
