build: 
	@GOOS="windows" GOARCH="amd64" go build -o bin/microsoft-oauth-demo.exe .

run: build
	@./bin/microsoft-oauth-demo
