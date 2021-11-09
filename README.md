# Varuh - Command line password manager

Password management done right for the Unix command line and the shell.

Table of Contents
=================

* [About](#about)
* [Building the code](#building-the-code)
* [Usage](#usage)
* [Databases](#databases)
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

## Copy an entry

To copy or clone an entry,

	$ $ varuh -C 1
	Cloned to new entry, id: 2

## Remove an entry

	$ varuh -R 1
	Entry with id 1 was removed from the database

It is an error if the id does not exist.

	$ varuh -R 3
	No entry with id 3 was found

## Switch to a new database

Once a database is active, creating another one automatically encrypts the current one and makes the new one the active database. The automatic encryption happens only if the configuration flag `auto_encrypt` is turned on (See section [Configuration](#config) below).

	$ varuh -I mysecrets
	Encrytping current database - /home/anand/mypasswds
	Password: 
	Password again: 
	Encryption complete.
	Created new database - mysecrets
	Updating active db path - /home/anand/mysecrets

The previous database is now encrypted with AES-256 cipher using the password. Please make sure you remember the password.

## Switch back to previous database

If you want to switch back to a previous database, you can use the `-U` option. The same process is repeated with the current database getting encrypted and the older one getting decrypted.

	$ varuh -U mypasswds
	Encrypting current active database - /home/anand/mysecrets
	Password: 
	Password again: 
	Encryption complete.
	Database /home/anand/mypasswds is encrypted, decrypting it
	Password: 
	Decryption complete.
	Switched active database successfully.
	
## Manual encryption and decryption

You can manually encrypt the current database using the `-e` option.

	$ varuh -e
	Password: 
	Password again: 
	Encryption complete.

Note that once you encrypt the active database, you cannot use the listings any more unless it is decrypted.

	$ varuh -l 2
	No decrypted active database found.

Manually decrypt the database using `-d` option.

	$ varuh -d mypasswds 
	Password: 
	Decryption complete.

Now the database is active again and you can see the listings.

	$ varuh -l 2
	=====================================================================
	ID: 2
	Title: My Blog Login
	User: myblog.name
	URL: http://meblog
	Password: *********
	Notes: Website uses Apache
	Modified: 2021-21-09 23:21:32
	=====================================================================

Listing and Searching
=====================

## List an entry using id

To list an entry using its id,

	$ varuh -l 8
	=====================================================================
	ID: 8
	Title: Google account
	User: anandpillai@alumni.iitm.ac.in
	URL: 
	Password: ***********
	Notes: 
	Modified: 2021-21-25 15:02:50
	=====================================================================

## To search an entry

An entry can be searched on its title, username, URL or notes. Search is case-insensitive.

	$ varuh -f google
	=====================================================================
	ID: 8
	Title: Google account
	User: anandpillai@alumni.iitm.ac.in
	URL: 
	Password: **********
	Notes: 
	Modified: 2021-21-25 15:02:50
	=====================================================================
	ID: 9
	Title: Google account
	User: xyz@gmail.com
	URL: 
	Password: ********
	Notes: 
	Modified: 2021-21-25 15:05:36
	=====================================================================
	ID: 10
	Title: Google account
	User: somethingaboutme@gmail.com
	URL: 
	Password: ***********
	Notes: 
	Modified: 2021-21-25 15:09:51
	=====================================================================

## To list all entries

To list all entries, use the option `-a`.

	$ varuh -a
	=====================================================================
	ID: 1
	Title: My Bank #1
	User: myusername1
	URL: https://mysuperbank1.com
	Password: ***********
	Notes: 
	Modified: 2021-21-15 15:40:29
	=====================================================================
	ID: 2
	Title: My Digital Locker #1
	User: mylockerusername
	URL: https://mysuperlocker1.com
	Password: **********
	Notes: 
	Modified: 2021-21-18 12:44:10
	=====================================================================
	ID: 3
	Title: My Bank Login #2
	User: mybankname2
	URL: https://myaveragebank.com
	Password: **********
	Notes: 
	Modified: 2021-21-19 14:16:33
	=====================================================================
	...

By default the listing is in ascending ID order. This can be changed in the configuration (see below).

## Turn on visible passwords

To turn on visible passwords, modify the configuration setting (see below) or use the `-s` flag.

## See current active database path

	$ varuh -p
	/home/anand/mypasswds

Configuration
=============

`Varuh` uses the standard [Free Desktop XDG Base Directory Spec](https://specifications.freedesktop.org/basedir-spec/basedir-spec-0.8.html) for storing its configuration in a JSON file. This usually translates to a folder name *.config/varuh* in your home directory on *nix systems.

The config file is named *config.json*. It looks as follows.

`{
	"active_db": "/home/anand/mypasswds",
	"auto_encrypt": true,
	"visible_passwords": false,
	"path": "/home/anand/.config/varuh/config.json",
	"list_order": "id,asc",
	"delimiter": "=",
	"color": "default",
	"bgcolor": "bgblack"
}
`
You can modify the following variables.

1. `auto_encrypt` - Set this to true to enable automatic encryption/decryption when switching databases. Otherwise you have to do this manually. The default is `true`.
2. `visible_passwords` - Set this to true to always show passwords in clear text in listings. Otherwise passwords are masked using asterisks. This can be overridden with the `-s` flag.
3. `list_order` - Ordering when using the `-a` option to view all listings. Supported fields are,
   * id - Uses the `ID` field.
   * timestamp - Uses the `Modified` timestamp field. Use this to show latest entries first.
   * title - Uses the `Title` field.
   * username - Uses the `User` field.

	Always specify this field as `<field>,<order>`. Supported `<order>` are `asc` and `desc`.
4. `delimiter` - This modifies the delimiter string when printing a listing. Only one character is allowed.
5. `color` - The foreground color of the text when printing listings.
6. `bgcolor` - The background color of the text when printing listings.

Visit this [gist](https://gist.github.com/abritinthebay/d80eb99b2726c83feb0d97eab95206c4) to see the supported color options. All color values must be in lower-case.

The fields `active_db` and `path` are used by the program for internal use. Please don't modify them!

License
=======

`Varuh` is licensed under the [GNU GPL V3](https://www.gnu.org/licenses/gpl-3.0.html) license. See the LICENSE file for details.
