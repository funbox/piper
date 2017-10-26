<p align="center"><a href="#piper">Piper</a> • <a href="#purpose">Purpose</a> • <a href="#installation">Installation</a> • <a href="#getting-started">Getting started</a> • <a href="#usage">Usage</a> • <a href="#license">License</a></p>

# Piper [![Go Report Card](https://goreportcard.com/badge/github.com/gongled/piper)](https://goreportcard.com/report/github.com/gongled/piper)

`piper` is a tiny and ease-to-use utility for log rotation for [12-factor apps that write their logs to stdout](https://12factor.net/logs). 

## Purpose

`piper` solves three problems of logging design:

- It supports log rotation by [POSIX signals](https://en.wikipedia.org/wiki/Signal_(IPC)#POSIX_signals), in contrast to similar tools.
- It rotates logs for apps that do not support log rotation out-of-box.
- It rotates logs for apps that cannot be stopped for some reasons.

## Installation

### From prebuilt package for RHEL7/CentOS7

```shell
[sudo] yum install -y http://yum.gongled.me/7/release/x86_64/gongled-release-7.4-0.el7.noarch.rpm
[sudo] yum install piper
```

### From the source code

Install dependencies using a Go package manager:

```shell
make deps
```

Build `piper`:

```shell
make all
```

## Getting started

Run `piper` to write stdin to `/var/log/program.log` file:

```
your_command_here | piper /var/log/program.log
```

You can also redirect output to `/dev/null` to suppress any output:

```
your_command_here | piper /var/log/program.log >/dev/null
```

Find PID and send USR1 signal to rotate log `/var/log/program.log` (see `signal(3)`):

```
kill -USR1 $PID
```

## Usage

```
Usage: piper {options} path

Options

  --no-color, -nc    Disable colored output
  --version, -v      Show information about version
  --help, -h         Show this help message

Examples

  piper /var/log/program.log
  Read info from the /dev/stdin and write to logging file
```

## License

MIT
