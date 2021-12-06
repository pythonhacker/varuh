# Varuh - Command line password manager

Password management done right for the Unix command line and the shell.

Table of Contents
=================

* [About](#about)
* [Install](#install)
* [Usage](#usage)
* [Encryption and Security](#encryption-and-security)
* [Databases](#databases)
* [Listing and Searching](#listing-and-searching)
* [Misc](#misc)
* [Export](#export)
* [Configuration](#configuration)
* [License](#license)
* [Feedback](#feedback)


About
=====

`Varuh` is a command line password manager that allows you to keep your passwords and other sensitive data using the power of the shell and Unix. It uses `sqlite` databases to store the information and encrypts it with symmetric encryption ciphers like [AES-256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) and [XChaCha20-Poly1305](https://www.cryptopp.com/wiki/XChaCha20) .

The name [Varuh](https://www.wordsense.eu/varuh/#Slovene) means *guardian* or *protector* in the Slovene language.

Varuh is inspired by [ylva](https://github.com/nrosvall/ylva) but it is full re-implementation - with some major changes in the key derivation functions and ciphers. It is written in `Go` and has been tested with Go versions 1.16 and 1.17 on Debian Linux (Antix). It should work on other versions of Linux and *BSD as well.

Install
=======

## Binary Release

If you are on a Debian or Debian derived system, you can directly download and install the latest version. Check out the [releases](https://github.com/pythonhacker/varuh/releases) page and use `dpkg` to install the binary.

	$ sudo dpkg -i varuh-${VERSION}_amd64.deb

The binary will be installed in `/usr/bin` folder.

## Building from Source

You need the [Go compiler](https://golang.org/dl/) to build the code. (This can be usually installed on \*nix machines by the native package managers like *apt-get*).

Install `make` by using your native package manager. Something like,

	$ sudo apt install make -y

should work.

Then,

	$ make
	go: downloading github.com/kirsle/configdir v0.0.0-20170128060238-e45d2f54772f
	go: downloading github.com/pborman/getopt/v2 v2.1.0
	go: downloading golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	go: downloading gorm.io/driver/sqlite v1.2.3
	...

	$ sudo make install
	Installing varuh...done

The binary will be installed in `/usr/local/bin` folder.


Usage
=====

	$ varuh -h


	SYNOPSIS

		varuh [options] [flags]

	OPTIONS

		EDIT/CREATE ACTIONS:

		  -A --add                        Add a new entry
		  -I --init            <path>     Initialize a new database
		  -R --remove          <id>       Remove an entry
		  -e --encrypt                    Encrypt the current database
		  -d --decrypt         <path>     Decrypt password database
		  -C --clone           <id>       Clone an entry
		  -U --use-db          <path>     Set as active database
		  -E --edit            <id>       Edit entry by id

		FIND/LIST ACTIONS:

		  -f --find            <term>     Search entries
		  -l --list-entry      <id>       List entry by id
		  -x --export          <filename> Export all entries to <filename>
		  -p --path                       Show current database path
		  -a --list-all                   List all entries in current database

		MISC ACTIONS:

		  -g --genpass         <length>   Generate password of given length

		HELP ACTIONS:

		  -h --help                       Print this help message and exit
		  -v --version                    Show version information and exit

		FLAGS:

		  -s --show                       Show passwords when listing entries
		  -c --copy                       Copy password to clipboard


	AUTHORS
		Copyright (C) 2021 Anand B Pillai <abpillai@gmail.com>


The command line flags are grouped into `Edit/Create`, `Find/List`, `Misc` and `Help` actions. The first group of actions allows you to work with password databases and perform create/edit as well as encrypt/decrypt actions. The second set of actions allows you to work with an active decrypted database and view/search/list entries.

Encryption and Security
=======================

Varuh gives the option of two symmetric ciphers - AES (default) and XChacha20-Poly1305.

AES is a block cipher supported with 256-bit key size for encryption and is the current standard for symmetric encryption ciphers.

XChacha20-Poly1305 is a stream cipher with a longer nonce (192 bits) which makes the cipher more resistant to timing attacks than AES-GCM. It also supports 256-bit key size.

The key derivation uses [Argon2](https://en.wikipedia.org/wiki/Argon2) with 32MB memory and 4 threads with a random cryptographic salt of 128 bit size for both ciphers.

Databases are created and decrypted with owner `rw` mode (0600). This makes sure the databases are read/write - able only by the owner.

When the `auto_encrypt` and `encrypt_on` flags are turned on, the database is always encrypted after an operation so the passwords remain in the clear in memory as well as in disk for a very short time. This increases the security of the data.

For maximum security, the default settings `auto_encrypt` and `encrypt_on` to true and `visible_passwords` to false is suggested.

Databases
=========

`Varuh` works with password databases. Each password database is a sqlite3 file. You can create any number of databases but at any given time there is only one active database which is in decrypted mode. When `auto_encrypt` is turned on (default), the program takes care of automatically encrypting and decrypting databases.

## Create a database

	$ varuh -I mypasswds
	Created new database - mypasswds
	Updating active db path - /home/anand/mypasswds

	$ ls -lt mypasswds 
	-rw------- 1 anand anand 8192 Nov  9 23:06 mypasswds

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
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 1
	Title: My Website Login
	User: mememe
	URL: http://mywebsite.name
	Password: ****************
	Notes: Website uses Nginx auth
	Modified: 2021-21-09 23:12:35
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

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
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 1
	Title: My Blog Login
	User: meblog
	URL: http://myblog.name
	Password: myblog123
	Notes: Website uses Apache
	Modified: 2021-21-09 23:15:29
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

(*-s* turns on visible passwords)

## Clone an entry

To clone (copy) an entry,

	$ $ varuh -C 1
	Cloned to new entry, id: 2

## Remove an entry

	$ varuh -R 1
	Entry with id 1 was removed from the database

It is an error if the id does not exist.

	$ varuh -R 3
	No entry with id 3 was found

## Switch to a new database

Once a database is active, creating another one automatically encrypts the current one and makes the new one the active database. The automatic encryption happens only if the configuration flag `auto_encrypt` is turned on (See section [Configuration](#configuration) below).

	$ varuh -I mysecrets
	Encrytping current database - /home/anand/mypasswds
	Password: 
	Password again: 
	Encryption complete.
	Created new database - mysecrets
	Updating active db path - /home/anand/mysecrets

The previous database is now encrypted with the configured block cipher using the password. Please make sure you remember the password.

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
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 2
	Title: My Blog Login
	User: myblog.name
	URL: http://meblog
	Password: *********
	Notes: Website uses Apache
	Modified: 2021-21-09 23:21:32
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

## Always on encryption

If the config param `encrypt_on` is set to `true` along with `auto_encrypt` (default), the program will keep encrypting the database after each action, whether it is an edit/listing action. In this mode, the decryption password is saved in memory and re-used for encryption to avoid too many password queries.

### Example

	$ varuh -f my -s
	Password: 
	Decryption complete.
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 2
	Title: MY LOCAL BANK
	User: banklogin
	URL: https://my.localbank.com
	Password: bankpass123
	Notes: 
	Modified: 2021-21-18 12:44:10
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	Encryption complete.

In this mode, your data is provided maximum safety as the database remains decrypted only for a short while on the disk while the data is being read and once done is encrypted back again.

Listing and Searching
=====================

## List an entry using id

To list an entry using its id,

	$ varuh -l 8
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 8
	Title: Google account
	User: anandpillai@alumni.iitm.ac.in
	URL: 
	Password: ***********
	Notes: 
	Modified: 2021-21-25 15:02:50
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

## To search an entry

An entry can be searched on its title, username, URL or notes. Search is case-insensitive.

	$ varuh -f google
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 8
	Title: Google account
	User: anandpillai@alumni.iitm.ac.in
	URL: 
	Password: **********
	Notes: 
	Modified: 2021-21-25 15:02:50
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 9
	Title: Google account
	User: xyz@gmail.com
	URL: 
	Password: ********
	Notes: 
	Modified: 2021-21-25 15:05:36
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 10
	Title: Google account
	User: somethingaboutme@gmail.com
	URL: 
	Password: ***********
	Notes: 
	Modified: 2021-21-25 15:09:51
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

## To list all entries

To list all entries, use the option `-a`.

	$ varuh -a
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 1
	Title: My Bank #1
	User: myusername1
	URL: https://mysuperbank1.com
	Password: ***********
	Notes: 
	Modified: 2021-21-15 15:40:29
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 2
	Title: My Digital Locker #1
	User: mylockerusername
	URL: https://mysuperlocker1.com
	Password: **********
	Notes: 
	Modified: 2021-21-18 12:44:10
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	ID: 3
	Title: My Bank Login #2
	User: mybankname2
	URL: https://myaveragebank.com
	Password: **********
	Notes: 
	Modified: 2021-21-19 14:16:33
	+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	...

By default the listing is in ascending ID order. This can be changed in the configuration (see below).

## Turn on visible passwords

To turn on visible passwords, modify the configuration setting (see below) or use the `-s` flag.

## Copy password to clipboard

To copy a password to clipboard, use the `-c` or `--copy` flag. This works *only if* the result for a listing is single. For example this will work when listing an entry by id or when a search results in a single hit. It *will not work* when listing all entries or when a search results in multiple hits.

This is useful to copy the password to a password input field in the browser for example.

## See current active database path

	$ varuh -p
	/home/anand/mypasswds

Export
======

`Varuh` allows to export password databases to the following formats.

1. `csv`
2. `markdown`
3. `html`
4. `pdf`

To export use the `-x` option. The type of file is automatically figured out from the filename extension.

	$ varuh -x passwds.csv
	!WARNING: Passwords are stored in plain-text!
	Exported 14 records to passwds.csv .
	Exported to passwds.csv.

	$ varuh -x passwds.html
	Exported to passwds.html.

PDF export is supported if `pandoc` is installed along with the required `pdflatex` packages. The following command (on `Debian` and derived systems) should install the required dependencies.

	$ sudo apt-get install pandoc texlive-latex-base texlive-fonts-recommended texlive-fonts-extra texlive-latex-extra texlive-xetex lmodern -y

Then,

	$ varuh -x passwds.pdf
	pdftk not found, PDF won't be secure!

	File passwds.pdf created without password.
	Exported to passwds.pdf.

PDF files are exported in landscape mode with 150 dpi and 600 columns. To avoid the data not fitting into one page the fields `Notes` and `URL` are not exported.

If `pdftk` is installed, the PDF files will be encrypted with an (optional) password.

	$ sudo apt-get install pdftk -y

	$ varuh -x passwds.pdf
	PDF Encryption Password: ******
	File passwds.pdf created without password.
	Added password to passwds.pdf.
	Exported to passwds.pdf.

Misc
====

The following miscellaneous actions are supported.

Generate a secure password of given length.

	$ varuh -g
	7nhga7tkk9LNafz

	By passing the `-c` option, the password is also copied to the clipboard.

	$ varuh -g 15 -c 
	yeXlLlk??IOsvL6
	Password copied to clipboard


Configuration
=============

`Varuh` uses the standard [Free Desktop XDG Base Directory Spec](https://specifications.freedesktop.org/basedir-spec/basedir-spec-0.8.html) for storing its configuration in a JSON file. This usually translates to a folder name *.config/varuh* in your home directory on *nix systems.

The config file is named *config.json*. It looks as follows.

	`{
		"active_db": "/home/anand/mypasswds",
		"cipher": "aes",
		"auto_encrypt": true,
		"visible_passwords": false,
		"encrypt_on": true,
		"path": "/home/anand/.config/varuh/config.json",
		"list_order": "id,asc",
		"delimiter": "+",
		"color": "default",
		"bgcolor": "bgblack"
	}
	`
You can modify the following variables.

1. `auto_encrypt` - Set this to true to enable automatic encryption/decryption when switching databases. Otherwise you have to do this manually. The default is `true`.
1. `cipher` - The block cipher to use. This is `aes` by default. To switch to `xchacha20-poly1305` set this to `xchacha`,`chacha` or `xchachapoly`.
1. `visible_passwords` - Set this to true to always show passwords in clear text in listings. Otherwise passwords are masked using asterisks. This can be overridden with the `-s` flag.
1. `encrypt_on` - Set this to true for the program to always encrypt the database after every action. This makes sure that the database is never sitting in the unencrypted form on the disk and increases the security.
1. `list_order` - Ordering when using the `-a` option to view all listings. Supported fields are,
   * `id` - Uses the `ID` field.
   * `timestamp` - Uses the `Modified` timestamp field. Use this to show latest entries first.
   * `title` - Uses the `Title` field.
   * `username` - Uses the `User` field.

	Always specify this configuration as `<field>,<order>`. Supported `<order>` values are `asc` and `desc`.
1. `delimiter` - This modifies the delimiter string when printing a listing. Only one character is allowed.
1. `color` - The foreground color of the text when printing listings.
1. `bgcolor` - The background color of the text when printing listings.

Visit this [gist](https://gist.github.com/abritinthebay/d80eb99b2726c83feb0d97eab95206c4) to see the supported color options. All color values must be in lower-case.

The fields `active_db` and `path` are for internal use. Suggest not to modify them.

License
=======

`Varuh` is licensed under the [GNU GPL V3](https://www.gnu.org/licenses/gpl-3.0.html) license. See the LICENSE file for details.

Feedback
========

Please send your valuable feedback and suggestions to my email available in the program's usage listing.
