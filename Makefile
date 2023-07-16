build:
	go build -o dist/gitwho main.go

run-authors:
	go run main.go authors

run-files:
	go run main.go files

run-ownership:
	go run main.go ownership

test:
	go test
	cd ownership && go test
