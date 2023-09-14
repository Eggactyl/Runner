# Eggactyl Old Runner
This is what Eggactyl used to use to run the whole thing. I have made some modifications to make it easier for other people to use it if they want to. This was made to circumvent a bug in Pterodactyl which makes it impossible to get a SIGINT in the application.  

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
- Golang 1.21 (You can probably use older versions, but I used this to build this version)
- Make

To build the app:
```bash
make
```

To install it: (If you want to?)
```bash
sudo make install
```