build:
	rm -rf npm/dist
	mkdir -p npm/dist

	cp npm/main.js npm/dist/

	@echo "Compiling for darwin-amd64"...
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -o npm/dist/gitwho-darwin-amd64 ../
	chmod +x npm/dist/gitwho-darwin-amd64

	@echo "Compiling for darwin-arm64 (M1,2)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -a -o npm/dist/gitwho-darwin-arm64 ../
	chmod +x npm/dist/gitwho-darwin-amd64

	@echo "Compiling for linux-amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o npm/dist/gitwho-linux-amd64 ../
	chmod +x npm/dist/gitwho-linux-amd64

	@echo "Compiling for linux-arm (Raspberry Pi)..."
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -a -o npm/dist/gitwho-linux-raspberry ../
	chmod +x npm/dist/gitwho-linux-amd64

	@echo "Compiling for windows-amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -o npm/dist/gitwho-windows-amd64.exe ../

publish-npm:
	git config user.email "flaviostutz@gmail.com"
	git config user.name "Flávio Stutz"
	npm version from-git --no-git-tag-version
	echo "//registry.npmjs.org/:_authToken=${NPM_ACCESS_TOKEN}" > .npmrc
	yarn publish
