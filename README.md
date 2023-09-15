# Eggactyl Runner
This is what Eggactyl uses to run the whole script. I have made some modifications to make it easier for other people to use it if they want to. This was made to circumvent a bug in Pterodactyl which makes it impossible to get a SIGINT in the application.  

This is probably not the best way to do this stuff, but it does work at least. If people do want to make commits, I am willing to merge them.

## How to Use:

To run a script:
```bash
/path/to/eggactyl_runner --script /path/to/script
```

If you want to include support information in the errors:
```bash
/path/to/eggactyl_runner --script /path/to/script --support-link https://support.example.com
```

## How to build

- UPX, you can download it [here](https://upx.github.io)
- Golang 1.21, you can download it [here](https://go.dev/dl/) (You can probably use older versions, but I used this to build this version)
- Make

### Install UPX
You can either use your package manager if possible:
```bash
sudo apt install upx
```

Or you can download it and install it manually:
```
wget https://github.com/upx/upx/releases/download/v4.1.0/upx-4.1.0-amd64_linux.tar.xz
sudo tar -xvf upx-4.1.0-amd64_linux.tar.xz -C /usr/local/bin upx-4.1.0-amd64_linux/upx
```

### Build the App
You can just run make:
```bash
make
```

Or you can use the go command to build it:
```bash
go build *.go -ldflags="-s -w" -trimpath -o runner *.go
```