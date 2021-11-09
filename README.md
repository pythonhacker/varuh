# Varuh - Command line password manager

Password management done right for the Unix command line and the shell.

Varuh is inspired by [ylva](https://github.com/nrosvall/ylva) and licensed under the GNU GPLv3 License.
It is written in `Go` and has been tested with Go versions 1.16 and 1.17 on Debian Linux (Antix). It should
work on other versions of Linux and *BSD as well.

Table of Contents
=================

* [About](#about)
* [Databases](#databases)
* [Encryption](#encryption)
* [Example Usage](#usage)
* [Configuration](#config)
* [License](#license)

About
=====

`Varuh` is a command line password manager that allows you to keep your passwords and other sensitive data using the power of the shell and Unix. It uses `sqlite` databases to store the information and encrypts it with [AES-256'(https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) block encryption.

The name [Varuh](https://www.wordsense.eu/varuh/#Slovene) means *guardian* or *protector* in the Slovene language.
