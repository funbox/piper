<p align="center"><a href="#piper">Piper</a> • <a href="#purpose">Purpose</a> • <a href="#installation">Installation</a> • <a href="#getting-started">Getting started</a> • <a href="#usage">Usage</a> • <a href="#license">License</a></p>

# Piper [![Go Report Card](https://goreportcard.com/badge/github.com/gongled/piper)](https://goreportcard.com/report/github.com/gongled/piper)

`piper` is a tiny and ease-to-use utility for log rotation for [12-factor apps that write their logs to stdout](https://12factor.net/logs). 

## Purpose

`piper` solves these problems of logging design:

- It supports log rotation by USR1 [POSIX signal](https://en.wikipedia.org/wiki/Signal_(IPC)#POSIX_signals), in contrast to similar tools.
- It rotates logs for apps that do not support log rotation out-of-box.
- It rotates logs for apps that cannot be stopped for some reasons.
- It rotates logs for apps by given duration.
- It rotates logs for apps by given size limit.

## Installation

### From the source code

Install dependencies using a Go package manager:

```shell
make deps
make all
```

## Getting started

Run `piper` to write stdin to `/var/log/program.log` file:

```
your_command_here | piper /var/log/program.log
```

Any output to stdout will be suppressed.

## Usage

```
Usage: piper {options} path

Options

  --size, -s size       Max file size
  --keep, -k number     Number of files to keep
  --age, -a interval    Interval of log rotation
  --timestamp, -t       Prepend timestamp to every entry
  --no-color, -nc       Disable colored output
  --version, -v         Show information about version
  --help, -h            Show this help message

Examples

  piper /var/log/program.log
  Read stdin and write entries to the logging file

  piper -t /var/log/program.log
  Prepend timestamp to every entry

  piper -s 5MB -k 10 /var/log/program.log
  Rotate logging file if it is reached 5M. Keep only 10 files

  piper -a 10m -k 5 /var/log/program.log
  Rotate logging file every 10 minute. Keep only 5 files
```

## License

Released under the MIT license (see [LICENSE](LICENSE))

[![Sponsored by FunBox](https://funbox.ru/badges/sponsored_by_funbox_grayscale.svg)](https://funbox.ru)
