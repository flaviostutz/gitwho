build:
	go build -o dist/gitwho main.go

run-changes:
	# go run main.go changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md --since "5 years ago" --until "3 years ago" --format top
	# go run main.go changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files contribs/src/test/resources/log4j.properties --since "4 years ago" --until "3 years ago"
	go run main.go changes --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files . --since "3 year ago" --until "2 year ago" --format full --show-mail true

run-ownership:
# gocv, orb, conductor
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch main --files .md
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files .*
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/nn/it4it-pipelines --branch no-build-stage --files .*
	go run main.go ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts" --when "1 years ago"

test:
	go test
	cd utils && go test
	cd ownership && go test

open-profile:
	go tool pprof -http=:8080 profile.pprof
