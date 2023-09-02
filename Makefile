SHELL := /bin/bash

build: build-npm-all

unit-tests:
	go test -cover -coverprofile=./ownership/coverage.out ./ownership
	go test -cover -coverprofile=./changes/coverage.out ./changes
	go test -cover -coverprofile=./utils/coverage.out ./utils
	go test -cover -coverprofile=./cli/coverage.out ./cli
	# make coverage

test: unit-tests

coverage:
	go tool cover -func ./ownership/coverage.out
	go tool cover -func ./changes/coverage.out 
	go tool cover -func ./utils/coverage.out
	go tool cover -func ./cli/coverage.out
	# open cover report on browser
	# go tool cover -html=./utils/coverage.out

benchmark:
	go test -bench . -benchmem -count 20

deploy: publish-npm-all

run-changes:
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --since "5 years ago" --until "3 years ago" --format full
	go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files ".*" --files-not "vendor" --since "30 days ago" --until "now" --format graph
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts$$" --since "15 days ago" --until "now" --format graph --authors "Flavio|Marcio|Niels|Gabriel" --verbose true

run-ownership:
# gocv, orb, conductor
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --format full
	go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files ".*" --files-not vendor --authors "Sebastiaan|Brian|Cory|Tõnis|Jana" --format graph
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files "shared" --files-not "" --when "now" --format graph

run-ownership-timeline:
# gocv, orb, conductor
	# go run ./ ownership-timeline --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --format full
	go run ./ ownership-timeline --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files test --files-not "vendor" --since="3 months ago" --until "now" --period "2 weeks" --authors "Sebastiaan|Brian|Cory|Tõnis|Jana" --format full
	# go run ./ ownership-timeline --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"
	# go run ./ ownership-timeline --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files "shared/libs/middy-middlewares" --files-not "" --since="3 months ago" --until "now" --period "2 weeks" --format graph


run-duplicates:
# gocv, orb, conductor
	# go run ./ duplicates --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --format full
	go run ./ duplicates --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files test --files-not "vendor" --format full --min-dup-lines 6
	# go run ./ duplicates --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"
	# go run ./ duplicates --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts$$" --files-not "" --when "now" --format full --min-dup-lines 6

publish-npm-all:
	@if [ "${NPM_ACCESS_TOKEN}" == "" ]; then \
		echo "ENV NPM_ACCESS_TOKEN is required"; \
		exit 1; \
	fi

	PACKAGE_DIR="npm/gitwho" make publish-npm-dir
	echo "Sleeping for 60s to avoid npm publish rate limiting..."
	sleep 60
	PACKAGE_DIR="npm/@gitwho/darwin-amd64" make publish-npm-dir
	echo "Sleeping for 60s to avoid npm publish rate limiting..."
	sleep 60
	PACKAGE_DIR="npm/@gitwho/darwin-arm64" make publish-npm-dir
	echo "Sleeping for 60s to avoid npm publish rate limiting..."
	sleep 60
	PACKAGE_DIR="npm/@gitwho/linux-amd64" make publish-npm-dir
	echo "Sleeping for 60s to avoid npm publish rate limiting..."
	sleep 60
	PACKAGE_DIR="npm/@gitwho/linux-arm64" make publish-npm-dir
	echo "Sleeping for 60s to avoid npm publish rate limiting..."
	sleep 60
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

	@echo "Publishing package to npmjs.org..."
	@echo "//registry.npmjs.org/:_authToken=${NPM_ACCESS_TOKEN}" > ${PACKAGE_DIR}/.npmrc
	cd ${PACKAGE_DIR} && yarn publish
	@echo "Done publishing"


open-profile:
	go tool pprof -http=:8080 profile.pprof
