build:
	go version
	go mod download

unit-tests:
	go test ./
	go test ./utils
	go test ./ownership
	go test ./changes

test: unit-tests

deploy:
	cd publish && make publish-npm

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
