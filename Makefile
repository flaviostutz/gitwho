build:
	go build -o dist/gitwho main.go

run-changes:
	go run main.go changes --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch master --files go/README.md --since "48 months ago" --until "47 months ago"

run-ownership:
# gocv, orb, conductor
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch master --files .md
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/moby --branch master --files .*
	# go run main.go ownership --repo /Users/flaviostutz/Documents/development/nn/it4it-pipelines --branch no-build-stage --files .*
	go run main.go ownership --repo /Users/flaviostutz/Documents/development/nn/mortgage-loan --branch master --files ".ts" --when "2 year ago"

test:
	go test
	cd utils && go test
	cd ownership && go test

open-profile:
	go tool pprof -http=:8080 profile.pprof
