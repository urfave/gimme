SHELL := bash
.DEFAULT_GOAL := all

GIMME_GENERATE := ./gimme-generate
KNOWN_BINARY_VERSIONS_FILES := \
	.testdata/binary-darwin \
	.testdata/binary-linux \
	.testdata/sample-binary-darwin \
	.testdata/sample-binary-linux

.PHONY: all
all: lint assert-copyright generate

.PHONY: clean
clean:
	$(RM) $(KNOWN_BINARY_VERSIONS_FILES) $(GIMME_GENERATE) .testdata/object-urls

.PHONY: lint
lint:
	git grep -l '^#!/usr/bin/env bash' | xargs shellcheck
	git grep -l '^#!/usr/bin/env bash' | xargs shfmt -i 0 -w

.PHONY: generate
generate: $(KNOWN_BINARY_VERSIONS_FILES)
	@true

$(GIMME_GENERATE): $(shell git ls-files '*.go')
	go build -o $@ ./internal/cmd/gimme-generate/

.PHONY: assert-copyright
assert-copyright:
	@diff -u \
		--label a/copyright/gimme \
		<(awk 'BEGIN { FS="="; } /^readonly GIMME_COPYRIGHT/ { gsub(/"/, "", $$2); print $$2 }' gimme) \
		--label b/copyright/LICENSE \
		<(awk '/^Copyright/ { print $$0 }' LICENSE)

.PHONY: assert-no-diff
assert-no-diff:
	git diff --exit-code && git diff --cached --exit-code

.PHONY: matrix
matrix: $(GIMME_GENERATE)
	$(GIMME_GENERATE) matrix-json

.PHONY: remove-object-urls
remove-object-urls:
	$(RM) .testdata/object-urls

.PHONY: force-update-versions
force-update-versions: remove-object-urls .testdata/object-urls
	@true

.PHONY: update-binary-versions
update-binary-versions: force-update-versions $(KNOWN_BINARY_VERSIONS_FILES)

.testdata/binary-%: .testdata/object-urls $(GIMME_GENERATE)
	$(GIMME_GENERATE) binary-list --os $* --from $^ >$@

.testdata/object-urls: $(GIMME_GENERATE)
	$(GIMME_GENERATE) go-links >$@

.testdata/sample-binary-%: .testdata/binary-% $(GIMME_GENERATE)
	$(GIMME_GENERATE) sample-binary-list --from $^ >$@
