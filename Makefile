build:
	go build -o dist/gitwho main.go

run-authors:
	-go run main.go authors
	git checkout main

run-files:
	-go run main.go files
	git checkout main

run-ownership:
	-go run main.go ownership
	git checkout main

test:
	go test
	cd utils && go test
	cd ownership && go test
