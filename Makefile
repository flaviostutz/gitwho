SHELL := /bin/bash

build: build-npm-all

unit-tests:
	go test ./
	go test ./utils
	go test ./ownership
	go test ./changes

test: unit-tests

deploy: publish-npm-all

run-changes:
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --since "5 years ago" --until "3 years ago" --format top
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files contribs/src/test/resources/log4j.properties --since "4 years ago" --until "3 years ago"
	go run ./ changes --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts$$" --since "9 months ago" --until "now" --format short

run-ownership:
# gocv, orb, conductor
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files .*
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/nn/it4it-pipelines --branch no-build-stage --files .*
	go run ./ ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts" --when "now"
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"


publish-npm-all:
	@if [ "${NPM_ACCESS_TOKEN}" == "" ]; then \
		echo "ENV NPM_ACCESS_TOKEN is required"; \
		exit 1; \
	fi

	PACKAGE_DIR="npm/gitwho" make publish-npm-dir
	PACKAGE_DIR="npm/@gitwho/darwin-amd64" make publish-npm-dir
	PACKAGE_DIR="npm/@gitwho/darwin-arm64" make publish-npm-dir
	PACKAGE_DIR="npm/@gitwho/linux-amd64" make publish-npm-dir
	PACKAGE_DIR="npm/@gitwho/linux-arm64" make publish-npm-dir
	PACKAGE_DIR="npm/@gitwho/windows-amd64" make publish-npm-dir

build-npm-all:
	@echo "Building binaries for all platforms..."
	OS=darwin ARCH=amd64 OUT_DIR="npm/@gitwho/darwin-amd64/dist" make build-arch-os
	OS=darwin ARCH=arm64 OUT_DIR="npm/@gitwho/darwin-arm64/dist" make build-arch-os
	OS=linux ARCH=amd64 OUT_DIR="npm/@gitwho/linux-amd64/dist" make build-arch-os
	OS=linux ARCH=arm64 OUT_DIR="npm/@gitwho/linux-arm64/dist" make build-arch-os
	OS=windows ARCH=amd64 OUT_DIR="npm/@gitwho/windows-amd64/dist" make build-arch-os
	@mkdir -p npm/gitwho/dist
	@cp npm/gitwho/gitwho npm/gitwho/dist/gitwho
	@echo "Build finished"

build-arch-os:
	@if [ "${ARCH}" == "" ]; then \
		echo "ENV ARCH is required"; \
		exit 1; \
	fi
	@if [ "${OS}" == "" ]; then \
		echo "ENV OS is required"; \
		exit 1; \
	fi
	
	@echo ""
	@echo "Compiling gitwho for ${OS}-${ARCH}"...

	rm -rf dist/${OS}-${ARCH}
	mkdir -p dist/${OS}-${ARCH}

	@go version
	go mod download

	GOOS=${OS} GOARCH=${ARCH} CGO_ENABLED=0 go build -a -o dist/${OS}-${ARCH}/gitwho
	chmod +x dist/${OS}-${ARCH}/gitwho

	@if [ "${OUT_DIR}" != "" ]; then \
		mkdir -p ${OUT_DIR}; \
		cp "dist/${OS}-${ARCH}/gitwho" "${OUT_DIR}/gitwho"; \
	fi
	@echo "Done compiling"


publish-npm-dir:
	@if [ "${NPM_ACCESS_TOKEN}" == "" ]; then \
		echo "ENV NPM_ACCESS_TOKEN is required"; \
		exit 1; \
	fi
	@if [ "${PACKAGE_DIR}" == "" ]; then \
		echo "ENV PACKAGE_DIR is required"; \
		exit 1; \
	fi

	@echo ""
	@echo "Preparing npm package ${PACKAGE_DIR}..."
	@if [ ! -f "${PACKAGE_DIR}/dist/gitwho" ]; then \
		echo "File '${PACKAGE_DIR}/dist/gitwho' not found. Forgot to run build?"; \
        exit 2; \
    fi

	VERSION=$$(npx -y monotag@1.5.1 latest); \
	sed -i -e "s/VERSION/$$VERSION/g" ${PACKAGE_DIR}/package.json
	rm ${PACKAGE_DIR}/package.json-e

	@echo "Publishing package to npmjs.org..."
	@echo "//registry.npmjs.org/:_authToken=${NPM_ACCESS_TOKEN}" > ${PACKAGE_DIR}/.npmrc
	cd ${PACKAGE_DIR} && yarn publish
	@echo "Done publishing"


open-profile:
	go tool pprof -http=:8080 profile.pprof
