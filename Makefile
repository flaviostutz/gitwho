build:
	rm -rf dist
	mkdir -p dist

	go version
	go mod download

	@echo "Compiling for darwin-amd64"...
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-darwin-amd64
	chmod +x dist/gitwho-darwin-amd64

	@echo "Compiling for darwin-arm64 (M1,2)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -a -o dist/gitwho-darwin-arm64
	chmod +x dist/gitwho-darwin-amd64

	@echo "Compiling for linux-amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-linux-amd64
	chmod +x dist/gitwho-linux-amd64

	@echo "Compiling for linux-arm (Raspberry Pi)..."
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -a -o dist/gitwho-linux-raspberry
	chmod +x dist/gitwho-linux-amd64

	@echo "Compiling for windows-amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-windows-amd64.exe

publish-npm: build
	if [ "${NPM_ACCESS_TOKEN}" == "" ]; then
		echo "NPM_ACCESS_TOKEN is a required env"
		exit 1
	fi

	rm -rf publish/npm/dist
	cp -r dist publish/npm/dist
	cp publish/npm/main.js publish/npm/dist/

	git config user.email "flaviostutz@gmail.com"
	git config user.name "Flávio Stutz"
	cd publish/npm && npm version from-git --no-git-tag-version
	echo "//registry.npmjs.org/:_authToken=${NPM_ACCESS_TOKEN}" > .npmrc
	cd publish/npm && yarn publish

unit-tests:
	go test ./
	go test ./utils
	go test ./ownership
	go test ./changes

test: unit-tests

deploy: publish-npm

run-changes:
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --since "5 years ago" --until "3 years ago" --format top
	# go run ./ changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files contribs/src/test/resources/log4j.properties --since "4 years ago" --until "3 years ago"
	go run ./ changes --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts$$" --since "3 year ago" --until "2 year ago" --format full --show-mail true

run-ownership:
# gocv, orb, conductor
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files .*
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/nn/it4it-pipelines --branch no-build-stage --files .*
	go run ./ ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts" --when "now"
	# go run ./ ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"

open-profile:
	go tool pprof -http=:8080 profile.pprof
