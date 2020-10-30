Pget - parallel file download client
=======

![test](https://github.com/Code-Hex/pget/workflows/test/badge.svg)
[![codecov](https://codecov.io/gh/Code-Hex/pget/branch/master/graph/badge.svg?token=jUVGnY7ZlG)](undefined)
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/pget)](https://goreportcard.com/report/github.com/Code-Hex/pget)
[![GitHub release](https://img.shields.io/github/release/Code-Hex/pget.svg)](https://github.com/Code-Hex/pget)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
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
    % pget -p 6 MIRROR1 MIRROR2 MIRROR3

If you have created such as this file

    cat list.txt
    MIRROR1
    MIRROR2
    MIRROR3

You can do this

    cat list.txt | pget -p 6

## Options

```
  Options:
  -h,  --help                   print usage and exit
  -v,  --version                display the version of pget and exit
  -p,  --procs <num>            split ratio to download file
  -o,  --output <filename>      output file to <filename>
  -d,  --target-dir <path>      path to the directory to save the downloaded file, filename will be taken from url
  -t,  --timeout <seconds>      timeout of checking request in seconds
  -u,  --user-agent <agent>     identify as <agent>
  -r,  --referer <referer>      identify as <referer>
  --check-update                check if there is update available
  --trace                       display detail error messages
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
