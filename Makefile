build:
	rm -rf dist
	mkdir dist

	@echo "Compile for darwin-amd64"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-darwin-amd64 ./cli
	chmod +x dist/gitwho-darwin-amd64
	@echo "Saved to dist/gitwho-darwin-amd64"

	@echo "Compile for darwin-arm64 (M1,2)"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -a -o dist/gitwho-darwin-arm64 ./cli
	chmod +x dist/gitwho-darwin-amd64
	@echo "Saved to dist/gitwho-darwin-amd64"

	@echo "Compile for linux-amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-linux-amd64 ./cli
	chmod +x dist/gitwho-linux-amd64
	@echo "Saved to dist/gitwho-linux-amd64"

	@echo "Compile for linux-arm (works on Raspberry)"
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -a -o dist/gitwho-linux-raspberry ./cli
	chmod +x dist/gitwho-linux-amd64
	@echo "Saved to dist/gitwho-linux-amd64"

	@echo "Compile for windows-amd64"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -o dist/gitwho-windows-amd64.exe ./cli
	@echo "Saved to dist/gitwho-windows-amd64"

unit-tests:
	cd cli && go test
	cd utils && go test
	cd ownership && go test
	cd changes && go test

test: unit-tests

run-changes:
	# go run ./cli changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --since "5 years ago" --until "3 years ago" --format top
	# go run ./cli changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files contribs/src/test/resources/log4j.properties --since "4 years ago" --until "3 years ago"
	go run ./cli changes --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts$$" --since "3 year ago" --until "2 year ago" --format full --show-mail true

run-ownership:
# gocv, orb, conductor
	# go run ./cli ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md
	# go run ./cli ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files .*
	# go run ./cli ownership --repo /Users/flaviostutz/Documents/development/nn/it4it-pipelines --branch no-build-stage --files .*
	go run ./cli ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts" --when "now"
	# go run ./cli ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/gitwho --branch main --files "." --when "now"

open-profile:
	go tool pprof -http=:8080 profile.pprof
