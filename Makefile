build:
	go build -o dist/gitwho main.go

run-authors:
	-go run main.go authors
	git checkout main

run-files:
	-go run main.go files
	git checkout main

run-ownership:
# gocv, orb, conductor
	go run main.go ownership --repo /Users/flaviostutz/Documents/development/flaviostutz/conductor --branch master --files ^/ui/src/api/.* --verbose true
	# git checkout main

test:
	go test
	cd utils && go test
	cd ownership && go test

open-profile:
	go tool pprof -http=:8080 profile.pprof
