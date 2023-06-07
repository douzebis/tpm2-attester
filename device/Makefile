# SPDX-License-Identifier: Apache-2.0

.PHONY: attest manifest

all: init attest manifests

init: src/init/main.go
	go build -o init src/init/main.go

attest:
	sed -e "s|ATTESTER_DEVICE_PATH|$$(pwd)|g" src/attest/main.go > src/attest/mainloc.go \
	  && go build -o attester src/attest/mainloc.go

manifests:
	./install-dev.sh -b firefox
	./install-dev.sh -b chromium