Pget - parallel file download client
=======

[![Build Status](https://travis-ci.org/Code-Hex/pget.svg?branch=master)](https://travis-ci.org/Code-Hex/pget)
[![Coverage Status](https://coveralls.io/repos/github/Code-Hex/pget/badge.svg?branch=master)](https://coveralls.io/github/Code-Hex/pget?branch=master)
[![License (GPL version 3)](https://img.shields.io/badge/license-GNU%20GPL%20version%203-blue.svg?style=flat-square)](http://opensource.org/licenses/GPL-3.0)

## Description

Download using a parallel requests

## Installation

    $ go get github.com/Code-Hex/pget/cmd/pget

## Synopsis

    % pget -p 6 URL

## Options

```
-h,  --help          print usage and exit
-v,  --version       display the version of pget and exit
-p,  --procs         split ratio to download file
-o,  --output        output file to FILENAME
--trace              display detail error messages
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

## Author

[codehex](https://twitter.com/CodeHex)
