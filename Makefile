# SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
# SPDX-License-Identifier: AGPL-3.0-only

GO_FILES != find . -name "*.go"

deskd: $(GO_FILES)
	go build -v -ldflags="-s" -o $@

.PHONY: test
test: $(GO_FILES)
	go test -v ./...

.PHONY: clean
clean:
	rm -f deskd

.PHONY: all
all: deskd
