all:
	rm -rf build
	mkdir build
	go build -ldflags="-s -w" -trimpath -o build/eggactyl_runner *.go
	upx --lzma --brute build/eggactyl_runner

clean:
	rm -rf build

install:
	ifeq ($(PREFIX),)
		PREFIX := /usr/local
	endif
	install -m 755 build/eggactyl_runner $(PREFIX)/bin