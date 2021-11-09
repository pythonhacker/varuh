# Varuh - Command line password manager

Password management done right for the Unix command line and the shell.

Table of Contents
=================

* [About](#about)
* [Building the code](#building-the-code)
* [Usage](#usage)
* [Databases](#databases)
* [Encryption](#encryption)
* [Listing and Searching](#listing-and-searching)
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

Usage
=====

	$ varuh -h

	SYNOPSIS

		varuh [options] [flags]

	OPTIONS

		EDIT/CREATE ACTIONS:

		  -U --use-db          <path> Set as active database
		  -E --edit            <id>   Edit entry by id
		  -e --encrypt                Encrypt the current database
		  -A --add                    Add a new entry
		  -I --init            <path> Initialize a new database
		  -d --decrypt         <path> Decrypt password database
		  -C --copy            <id>   Copy an entry
		  -R --remove          <id>   Remove an entry

		FIND/LIST ACTIONS:

		  -l --list-entry      <id>   List entry by id
		  -f --find            <term> Search entries
		  -p --path                   Show current database path
		  -a --list-all               List all entries in current database

		HELP ACTIONS:

		  -h --help                   Print this help message and exit
		  -v --version                Show version information and exit

		FLAGS:

		  -s --show                   Show passwords when listing entries


	AUTHORS
		Copyright (C) 2021 Anand B Pillai <anandpillai@alumni.iitm.ac.in>

The command line flags are grouped into `Edit/Create`, `Find/List` and `Help` actions. The first group of actions allows you to work with password databases and perform create/edit as well as encrypt/decrypt actions. The second set of actions allows you to work with an active decrypted database and view/search/list entries.

Databases
=========

`Varuh` works with password databases. Each password database is an sqlite3 file. You can create any number of databases but at any given time there is only one active database which is in decrypted mode. When `auto_encrypt` is turned on (default), the program takes care of automatically encrypting and decrypting databases.

## Create a database

	$ varuh -I mypasswds
	Created new database - mypasswds
	Updating active db path - /home/anand/mypasswds

	$ ls -lt mypasswds 
	-rw-r--r-- 1 anand anand 8192 Nov  9 23:06 mypasswds

The password database is created and is active now. You can start adding entries to it.

## Add an entry

	$ varuh -A
	Title: My Website Login
	URL: mywebsite.name
	Username: mememe
	Password (enter to generate new): 
	Generating password ...done
	Notes: Website uses Nginx auth
	Created new entry with id: 1

You can now list the entry with one of the list options.

	$ varuh -l 1
	=====================================================================
	ID: 1
	Title: My Website Login
	User: mememe
	URL: http://mywebsite.name
	Password: ****************
	Notes: Website uses Nginx auth
	Modified: 2021-21-09 23:12:35
	=====================================================================

For more on listing see the [Listing and Searching](#listing-and-searching) section below.

## Edit an entry

	$ varuh -E 1
	Current Title: My Website Login
	New Title: My Blog Login
	Current URL: http://mywebsite.name
	New URL: myblog.name
	Current Username: mememe
	New Username: meblog
	Current Password: lTzC2z9kRppnYsYl
	New Password ([y/Y] to generate new, enter will keep old one): 
	Current Notes: Website uses Nginx auth
	New Notes: Website uses Apache
	Updated entry.

	$ varuh -l 1 -s
	=====================================================================
	ID: 1
	Title: My Blog Login
	User: meblog
	URL: http://myblog.name
	Password: myblog123
	Notes: Website uses Apache
	Modified: 2021-21-09 23:15:29
	=====================================================================

(*-s* turns on visible passwords)

## Remove an entry

	$ varuh -R 1
	Entry with id 1 was removed from the database

It is an error if the id does not exist.

	$ varuh -R 2
	No entry with id 2 was found

