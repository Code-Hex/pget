Pget - The fastest file download client
=======

[![.github/workflows/main.yaml](https://github.com/Code-Hex/pget/actions/workflows/main.yaml/badge.svg)](https://github.com/Code-Hex/pget/actions/workflows/main.yaml)
[![codecov](https://codecov.io/gh/Code-Hex/pget/branch/master/graph/badge.svg?token=jUVGnY7ZlG)](undefined)
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/pget)](https://goreportcard.com/report/github.com/Code-Hex/pget)
[![GitHub release](https://img.shields.io/github/release/Code-Hex/pget.svg)](https://github.com/Code-Hex/pget)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
## Description

Multi-Connection Download using parallel requests.

- Fast
- Resumable
- Cross-compiled (windows, linux, macOS)

This is an example to download [linux kernel](https://www.kernel.org/). It will be finished between 15s.

![pget](https://user-images.githubusercontent.com/6500104/147878414-321c57ad-cff2-40f3-b2a4-12c30ff1363f.gif)


## Disclaimer

This program comes with no warranty. You must use this program at your own risk.

### Note

- Using a large number of connections to a single URL can lead to DOS attacks.
- The case is increasing that if you use multiple connections to 1 URL does not increase the download speed with the spread of CDNs.
  - I recommend to use multiple mirrors simultaneously for faster downloads (And the number of connections is 1 for each).

## Installation

### Homebrew

	brew update
	brew install Code-Hex/tap/pget

### Go

    $ go install github.com/Code-Hex/pget/cmd/pget@latest

## Synopsis

This example will be used 2 connections per URL.

    $ pget -p 2 MIRROR1 MIRROR2 MIRROR3

If you have created such as this file

    cat list.txt
    MIRROR1
    MIRROR2
    MIRROR3

You can do this

    cat list.txt | pget -p 2

## Options

```
  Options:
  -h,  --help                   print usage and exit
  -p,  --procs <num>            the number of connections for a single URL (default 1)
  -o,  --output <filename>      output file to <filename>
  -t,  --timeout <seconds>      timeout of checking request in seconds
  -u,  --user-agent <agent>     identify as <agent>
  -r,  --referer <referer>      identify as <referer>
  --check-update                check if there is update available
  --trace                       display detail error messages
```

## Pget vs Wget

URL: https://mirror.internet.asn.au/pub/ubuntu/releases/21.10/ubuntu-21.10-desktop-amd64.iso

Using
```
time wget https://mirror.internet.asn.au/pub/ubuntu/releases/21.10/ubuntu-21.10-desktop-amd64.iso
time pget -p 6 https://mirror.internet.asn.au/pub/ubuntu/releases/21.10/ubuntu-21.10-desktop-amd64.iso
```
Results

```
wget   3.92s user 23.52s system 3% cpu 13:35.24 total
pget -p 6   10.54s user 34.52s system 25% cpu 2:56.93 total
```

`wget` 13:35.24 total, `pget -p 6` **2:56.93 total (6x faster)**

## Binary

You can download from [here](https://github.com/Code-Hex/pget/releases)

## Author

[codehex](https://twitter.com/CodeHex)
