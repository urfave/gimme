SHELL := bash
UNAME := $(shell uname)
VERSION := $(shell git describe --always --tags)
.DEFAULT_GOAL := all

# Affects sorting for CONTRIBUTORS file; unfortunately these are not
# totally names (standards opaque IIRC) but this should work for us.
LC_COLLATE := en_US.UTF-8
# Alas, macOS collation is broken and generates spurious differences.

SED ?= sed
SORT ?= sort
UNIQ ?= uniq
ifeq ($(UNAME), Darwin)
	SED := gsed
	SORT := gsort
	UNIQ := guniq
endif

ifeq "$(shell $(SORT) --version-sort </dev/null >/dev/null 2>&1 || echo no)" "no"
	_ := $(warning "$(SORT) --version-sort not available, falling back to shell")
	REV_VERSION_SORT := $(SED) -E 's/\.([0-9](\.|$$))/.00\1/g; s/\.([0-9][0-9](\.|$$))/.0\1/g' | $(SORT) --general-numeric-sort -r | $(SED) 's/\.00*/./g'
else
	REV_VERSION_SORT := $(SORT) --version-sort -r
endif

SED_STRIP_COMMENTS ?= $(SED) -n -e '/^[^\#]/p'

KNOWN_BINARY_VERSIONS_FILES := \
	.testdata/binary-darwin \
	.testdata/binary-linux \
	.testdata/sample-binary-darwin \
	.testdata/sample-binary-linux

.PHONY: all
all: lint CONTRIBUTORS assert-copyright $(KNOWN_BINARY_VERSIONS_FILES)

.PHONY: clean
clean:
	$(RM) $(KNOWN_BINARY_VERSIONS_FILES) .testdata/object-urls
ifeq ($(UNAME), Darwin)
	$(warning Not deleting CONTRIBUTORS on macOS, locale sorting is broken)
else
	$(RM) CONTRIBUTORS
endif

.PHONY: lint
lint:
	git grep -l '^#!/usr/bin/env bash' | xargs shellcheck
	git grep -l '^#!/usr/bin/env bash' | xargs shfmt -i 0 -w

.PHONY: assert-copyright
assert-copyright:
	@diff -u \
		--label a/copyright/gimme \
		<(awk 'BEGIN { FS="="; } /^readonly GIMME_COPYRIGHT/ { gsub(/"/, "", $$2); print $$2 }' gimme) \
		--label b/copyright/LICENSE \
		<(awk '/^Copyright/ { print $$0 }' LICENSE)

.PHONY: matrix
matrix:
	go run ./generate-matrix-json.go

.PHONY: remove-object-urls
remove-object-urls:
	$(RM) .testdata/object-urls

.PHONY: force-update-versions
force-update-versions: remove-object-urls .testdata/object-urls
	@true

.PHONY: update-binary-versions
update-binary-versions: force-update-versions $(KNOWN_BINARY_VERSIONS_FILES)

.testdata/binary-%: .testdata/object-urls
	$(RM) $@
	cat .testdata/stubheader-all > $@
	cat $< | \
		grep -E "$(lastword $(subst -, ,$@)).*tar\.gz$$" | \
		awk -F/ '{ print $$5 }' | \
		$(SED) "s/\.$(lastword $(subst -, ,$@)).*//;s/^go//" | \
		$(SORT) -r | $(UNIQ) >> $@

.testdata/object-urls:
	./fetch-object-urls >$@

.testdata/sample-binary-%: .testdata/binary-%
	$(RM) $@
	cat .testdata/stubheader-sample > $@
	for prefix in $$($(SED_STRIP_COMMENTS) $< | $(SED) -En 's/^([0-9]+\.[0-9]+)(\..*)?$$/\1/p' | $(REV_VERSION_SORT) | $(UNIQ)) ; do \
		grep "^$${prefix}" $< | grep -vE 'rc|beta' | $(REV_VERSION_SORT) | head -1 >> $@ ; \
	done

CONTRIBUTORS:
ifeq ($(UNAME), Darwin)
	$(error macOS appears to have broken collation and will make spurious differences)
endif
	@echo 'gimme was built by these wonderful humans:' >$@
	@git log --format=%an | $(SORT) | $(UNIQ) | $(SED) 's/^/- /' >>$@
