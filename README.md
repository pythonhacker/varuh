# Varuh - Command line password manager

Password management done right for the Unix command line and the shell.

Table of Contents
=================

* [About](#about)
* [Building the code](#building-the-code)
* [Databases](#databases)
* [Encryption](#encryption)
* [Example Usage](#usage)
* [Configuration](#config)
* [License](#license)

About
=====

`Varuh` is a command line password manager that allows you to keep your passwords and other sensitive data using the power of the shell and Unix. It uses `sqlite` databases to store the information and encrypts it with [AES-256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) block encryption.

The name [Varuh](https://www.wordsense.eu/varuh/#Slovene) means *guardian* or *protector* in the Slovene language.

Varuh is inspired by [ylva](https://github.com/nrosvall/ylva). It is written in `Go` and has been tested with Go versions 1.16 and 1.17 on Debian Linux (Antix). It should work on other versions of Linux and *BSD as well.

Building the code
=================

You need the [Go compiler](https://golang.org/dl/) to build the code. (This can be usually installed on \*nix machines by the native package managers like *apt-get*).

If you have `make` installed,

	$ make
	go: downloading github.com/kirsle/configdir v0.0.0-20170128060238-e45d2f54772f
	go: downloading github.com/pborman/getopt/v2 v2.1.0
	go: downloading golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	go: downloading gorm.io/driver/sqlite v1.2.3
	...

	$ sudo make install
	Installing varuh...done

The binary will be installed in `/usr/local/bin` folder.

If you don't have `make`,

	$ go mod tidy
	$ go build -o varuh *.go
	$ sudo cp varuh /usr/local/bin/


