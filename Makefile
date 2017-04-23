.PHONY: default clean clobber build install preinstall test run
.PHONY: killer lint test unit-test coverage covered setup

default: install

MYAPP := apidCRUD
VENDOR_DIR := github.com/30x/$(MYAPP)/vendor
COV_DIR := cov
COV_FILE := $(COV_DIR)/covdata.out
LOG_DIR := logs
SQLITE_PKG := github.com/mattn/go-sqlite3

clean:
	go clean
	/bin/rm -rf $(LOG_DIR)
	mkdir -p $(LOG_DIR)
	/bin/rm -rf $(COV_DIR)
	mkdir -p $(COV_DIR)

clobber: clean
	/bin/rm -rf ./vendor

get:
	[ -d ./vendor ] \
	|| glide install

build:
	time go $@

setup:
	mkdir -p $(LOG_DIR) $(COV_DIR)

# install this separately to speed up compilations.  thanks to Scott Ganyo.
preinstall: get
	[ -d $(VENDOR_DIR)/$(SQLITE_PKG) ] \
	|| go install $(VENDOR_DIR)/$(SQLITE_PKG)

install: setup preinstall
	go $@ ./cmd/$(MYAPP)

run: install
	./runner.sh

killer:
	pkill -f $(MYAPP)

test: unit-test

unit-test:
	go test -coverprofile=$(COV_FILE) \
	| tee $(LOG_DIR)/$@.out
	go tool cover -func=$(COV_FILE) \
	> $(LOG_DIR)/cover-func.out

func-test:
	./tester.sh | tee $(LOG_DIR)/$@.out

# obsolete
coverage:
	./cover.sh | tee $(LOG_DIR)/$@.out

covered:
	./tested_funcs.sh | sort | tee $(LOG_DIR)/$@.out
	./uncovered.sh | tee $(LOG_DIR)/uncovered.out

lint:
	gometalinter.v1 --sort=path -e "don't use underscores" \
	| tee $(LOG_DIR)/$@.out
