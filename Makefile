SHELL := bash
.DEFAULT_GOAL := all

GIMME_GENERATE := ./gimme-generate

TAG_VERSION ?= notset

.PHONY: all
all: lint assert-copyright generate

.PHONY: clean
clean:
	$(RM) $(GIMME_GENERATE) .testdata/known-versions.txt .testdata/sample-versions.txt

.PHONY: lint
lint:
	git grep -l '^#!/usr/bin/env bash' | xargs shellcheck
	git grep -l '^#!/usr/bin/env bash' | xargs shfmt -i 0 -w

.PHONY: generate
generate: .testdata/sample-versions.txt
	@true

.PHONY: tag
tag: .assert-tag-version-defined .assert-tag-matches-source
	git tag -s -a -m 'Release $(TAG_VERSION)' '$(TAG_VERSION)' $(TAG_REF)

.PHONY: .assert-tag-version-defined
.assert-tag-version-defined:
ifeq ($(TAG_VERSION), notset)
	$(error TAG_VERSION must be set)
endif

.PHONY: .assert-tag-matches-source
.assert-tag-matches-source:
	@diff -u \
		--label a/version/gimme \
		<(awk 'BEGIN { FS="="; } /^readonly GIMME_VERSION/ { gsub(/"/, "", $$2); print $$2 }' gimme) \
		--label b/version/TAG_VERSION \
		<(echo '$(TAG_VERSION)')

$(GIMME_GENERATE): $(shell git ls-files '*.go') internal/sample-stub-header
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
	$(GIMME_GENERATE) matrix-json --from .testdata/sample-versions.txt

.PHONY: remove-known-versions
remove-known-versions:
	$(RM) .testdata/known-versions.txt

.PHONY: force-update-versions
force-update-versions: remove-known-versions .testdata/known-versions.txt
	@true

.PHONY: update-versions
update-versions: force-update-versions .testdata/sample-versions.txt

.testdata/known-versions.txt: $(GIMME_GENERATE)
	GIMME_VERSION_PREFIX=$(CURDIR)/.testdata ./gimme -k -l >/dev/null

.testdata/sample-versions.txt: .testdata/known-versions.txt $(GIMME_GENERATE)
	$(GIMME_GENERATE) sample-versions --from $< >$@
