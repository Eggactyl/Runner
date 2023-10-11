all:
	mkdir -p build
	go build -ldflags="-s -w" -trimpath -o build/eggactyl_runner *.go

all_arch:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o build/eggactyl_runner_amd64 *.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o build/eggactyl_runner_arm64 *.go

clean:
	rm -rf build