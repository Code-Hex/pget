Pget - parallel file download client
=======

[![Build Status](https://travis-ci.org/Code-Hex/pget.svg?branch=master)](https://travis-ci.org/Code-Hex/pget)
[![Coverage Status](https://coveralls.io/repos/github/Code-Hex/pget/badge.svg?branch=master)](https://coveralls.io/github/Code-Hex/pget?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/pget)](https://goreportcard.com/report/github.com/Code-Hex/pget)
[![GitHub release](https://img.shields.io/github/release/Code-Hex/pget.svg)](https://github.com/Code-Hex/pget)
[![License (GPL version 3)](https://img.shields.io/badge/license-GNU%20GPL%20version%203-blue.svg?style=flat-square)](http://opensource.org/licenses/GPL-3.0)
## Description

Download using a parallel requests

[![asciicast](https://asciinema.org/a/a505e9fpkdpd7ycefyjs3h5bb.png)](https://asciinema.org/a/a505e9fpkdpd7ycefyjs3h5bb)

## Installation

### Homebrew

	brew tap Code-Hex/pget
	brew install pget

### go get
Install

    $ go get github.com/Code-Hex/pget/cmd/pget

Update

    $ go get -u github.com/Code-Hex/pget/cmd/pget

## Synopsis

    % pget -p 6 URL

## Options

```
  Options:
  -h,  --help                      print usage and exit
  -v,  --version                   display the version of pget and exit
  -p,  --procs <num>               split ratio to download file
  -o,  --output <PATH|FILENAME>    output file to PATH or FILENAME
  -t,  --timeout <seconds>         timeout of checking request in seconds
  --check-update                   check if there is update available
  --trace                          display detail error messages
```

## Pget vs Wget

URL: http://ubuntutym2.u-toyama.ac.jp/ubuntu/16.04/ubuntu-16.04-desktop-amd64.iso

Using
```
time wget http://ubuntutym2.u-toyama.ac.jp/ubuntu/16.04/ubuntu-16.04-desktop-amd64.iso
time pget -p 6 http://ubuntutym2.u-toyama.ac.jp/ubuntu/16.04/ubuntu-16.04-desktop-amd64.iso
```
Results

```
wget   3.92s user 23.52s system 3% cpu 13:35.24 total
pget -p 6   10.54s user 34.52s system 25% cpu 2:56.93 total
```

## Binary
You can download from [here](https://github.com/Code-Hex/pget/releases)

## Author

[codehex](https://twitter.com/CodeHex)
