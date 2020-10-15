build:
	go build -o app.wasm

compile-wasm:
	echo "Staring compile for wasm"
	set GOOS=js 
	set GOARCH=wasm 
	go build -o web\app.wasm app\main.go

compile-handler:
	echo "Staring compile for handler
	go build -o main.exe main.go

run: 
	set GOOS=js
	set GOARCH=wasm
	go run main.go

clean:
	@go clean
	@-rm app.wasm