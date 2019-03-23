# rover
                                                                  _.=
    .----.-----.--.--.-----.----.                            +=====|
    |   _|  _  |  |  |  -__|   _|          ----========      |-___-|
    |__| |_____|\___/|_____|__|    ----================   @--|@---@|--@


> rover |ˈrōvər| n. a vehicle for driving over rough terrain, especially one driven by remote control over extraterrestrial terrain

[![Go Report Card](https://goreportcard.com/badge/github.com/brianshumate/rover)](https://goreportcard.com/report/github.com/brianshumate/rover)

## What?

`rover` is a CLI tool that explores systems and stores what it finds in plain text files which can optionally be archived and uploaded to remote storage.

It is inspired by tools like Apple's `sysdiagnose` and my friend Brent's [debug-ninja](https://github.com/fprimex/debug-ninja) utilities.

`rover` gathers important factoids from a system to paint a detailed picture of the current operating environment for engineers and operators curious about such things.

The general types of information `rover` gathers include:

- Operating system command output
- Operating system logging
- Application or service specific command output
- Application or service specific logging

All of the stored information can then be packaged up into a zip file named for the host, and shared however you prefer. Currently, `rover` directly supports shipping the zip file to an S3 bucket as well.

> See the **Internals** section for a more detailed breakdown of the specific commands that `rover` will attempt to execute on a given platform

## Why?

To assist with troubleshooting of systems and help make the process efficient, reliable, and repeatable.

It reliably and quickly gathers a wealth of targeted operational configuration items and metrics which are often extremely helpful for troubleshooting systems.

## How?

A CLI tool called `rover` that is written in [Go](https://golang.org/)

`rover` is a relatively small (16MB) static binary that is specifically aimed at systems running FreeBSD, Linux, or macOS for the time being.

### Running

If you have a Go environment, you can install and run the `rover` command like this:

```
$ go get github.com/brianshumate/rover
$ rover
Usage: rover [--version] [--help] <command> [<args>]

Available commands are:
    archive    Archive rover data into zip file
    consul     Execute Consul related commands and store output
    info       Output basic system factoids
    nomad      Execute Nomad related commands and store output
    system     Execute system commands and store output
    upload     Uploads rover archive file to S3 bucket
    vault      Execute Vault related commands and store output
```

### Building

Otherwise, you can consult the [Go documentation for the Go tools](https://golang.org/doc/install#install) for your platform (be sure it is one of the previously mentioned supported OS), and ***go*** from there!

If you change into the `$GOPATH/src/github.com/brianshumate/rover` directory, you can build `rover` binaries for different platforms by typing `make`.

Binaries are located in the following subdirectories of `pkg/` after a successful build:

```
├── pkg
│   ├── darwin-amd64
│   │   └── rover
│   ├── freebsd-amd64
│   │   └── rover
│   └── linux-amd64
│       └── rover
```

#### Development Build

If you'd prefer to make a development build for just your own host architecture and OS, you can use `make dev` and the `rover` binary will also be located in `./bin` after successful compiliation.

#### Running a Local Development Build

It's a single binary, `rover` with built in help:

```
$ make dev
...
$ ./bin/rover
Usage: rover [--version] [--help] <command> [<args>]

Available commands are:
    archive    Archive rover data into zip file
    consul     Execute Consul related commands and store output
    info       Output basic system factoids
    nomad      Execute Nomad related commands and store output
    system     Execute system commands and store output
    upload     Uploads rover archive file to S3 bucket
    vault      Execute Vault related commands and store output
```

## Configuration

Currently all configuration is specified with flags at runtime or set as environment variables.

For detailed help, including available flags, use `rover <command> --help`.

Environment variables are documented in their relevant sections.

## Commands

`rover` is primarily concerned with gathering useful operational intelligence from an environment. It can also currently pack up that intelligence, and ship it to an S3 bucket.

Here are the current commands and their details.

### archive

The `rover archive` command is used once you have used other commands to gather data.

It expects a directory in the present working directory with the same name as the system hostname, where `rover` has previously stored command outputs which it will compress into a zip archive named in this format:

```
rover-[hostname]-[date-time].zip
```

Example:

```
$ rover archive
Data archived in rover-penguin-20190322202232.zip
```

### consul

The `rover consul` command uses both OS tools and the `consul` binary (if found in PATH) to gather intelligence about and from the perspective of the local Consul agent.

The following unauthenticated commands are used:

- `consul version`

The following commands requiring a token with sufficient capabilities are used:

- `consul info`
- `consul members`
- `consul operator raft list-peers`

If the `CONSUL_HTTP_TOKEN` environment variable is set to the value of a token with sufficient privileges, that token value will be used for the authenticated requests.

In addition to these commands, `rover consul` checks and records some details from the process table on Linux hosts:

- `/proc/$(pidof consul)/limits`
- `/proc/$(pidof consul)/status`
- `/proc/$(pidof consul)/fd`

Finally, the `rover consul` command attempts to read and store Consul related entries from the system logs using the following sources:

- `/var/log/syslog`
- `/var/log/messages`
- journald

Example:

```
$ rover consul
Executed Consul related commands and stored output
```

### info

The `info` command presents an overview of some basic details `rover` has learned about the system it is executed on. The output will resemble the following example:

```
$ rover info
Basic factoids about this system:

OS:            linux
Architecture:  amd64
Date/Time:     Fri Mar 22 20:19:43 2019

Active running versions:

Consul version:  v1.4.3
Vault version:   v1.1.0
```

### nomad

The `rover nomad` command uses both OS tools and the `nomad` binary (if found in PATH) to gather intelligence about and from the perspective of the local Nomad agent.

The following commands are used:

- `nomad version`
- `nomad status`
- `nomad operator raft list-peers`

In addition to these commands, `rover nomad` checks and records some details from the process table on Linux hosts:

- `/proc/$(pidof nomad)/limits`
- `/proc/$(pidof nomad)/status`
- `/proc/$(pidof nomad)/fd`

Finally, the `rover nomad` command attempts to read and store Nomad related entries from the system logs using the following sources:

- `/var/log/syslog`
- `/var/log/messages`
- journald

Example:

```
$ rover nomad
Executed Nomad related commands and stored output
```

### system

The `rover system` command does a bit of work to determine something about the system it's been executed on, then proceeds to execute several commands (as described in the **Internals** section) and saves the output of the commands to simple text files.

The commands used for this are documented in detail within the **System Commands** section.

Example:

```
$ rover system
Executed system related commands and stored output
```

### upload

The `rover upload` command is used to upload an archive to an AWS S3 bucket.

The command has one required flag, `-file=` for specifying the file to upload; you must also set the following environment variables to use `rover upload`:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_BUCKET`
- `AWS_REGION`

Optionally, specify a bucket prefix with:

- `AWS_PREFIX`

Example:

```
$ rover system && rover archive
Executed system commands and stored output
Data archived in rover-penguin-20190322202232.zip

$ rover upload -file=rover-penguin-20190322202232.zip
Success! Uploaded rover-penguin-20190322202232.zip
```

### vault

The `rover vault` command uses both OS tools and the `vault` binary (if found in PATH) to gather intelligence about and from the perspective of the local Vault server.

The following unauthenticated commands are used:

- `vault version`
- `vault status`

The following commands requiring a token with sufficient capabilities are used:

- `vault audit list`
- `vault auth list`
- `vault secrets list`

If the `VAULT_TOKEN` environment variable is set to the value of a token with sufficient privileges, that token value will be used for the authenticated requests.

In addition to these commands, `rover vault` checks and records some details from the process table on Linux hosts:

- `/proc/$(pidof vault)/limits`
- `/proc/$(pidof vault)/status`
- `/proc/$(pidof vault)/fd`

Finally, the `rover vault` command attempts to read and store Vault related entries from the system logs using the following sources:

- `/var/log/syslog`
- `/var/log/messages`
- journald

Example:

```
$ rover vault
Executed Vault related commands and stored output
```

### Command Combinations

You can chain commands together to build a zip file with your desired contents like this:

```
$ rover consul && \
  rover vault && \
  rover system && \
  rover archive
Executed Consul related commands and stored output
Executed Vault related commands and stored output
Executed system related commands and stored output
Data archived in rover-penguin-20190322202232.zip
```

Investigation into simplified meta commands and easy one-liners is also on the roadmap.

## Internals

You don't need to know all of this to use `rover`, but it is documented here for ease of reference by those who'd like more detail without reading the source code.

### System Commands

By default, `rover` executes operating commands and stores both standard output and standard error into a plain text file. If the command is missing or requires additional privileges to execute, this will be captured in both the stored output and the rover log.

You can archive the files into a zip file with `rover archive`. Currently, this creates a zip file named with a timestamp and the system's hostname if it can be determined, resulting in a filename like this: `rover-waves-20170826135317.zip`.

Here is an example tree of an uncompressed zip file which was run to store Consul, system, and Vault information on macOS:

```
└── waves
    ├── consul
    │   ├── consul_info.txt
    │   ├── consul_members.txt
    │   ├── consul_operator_raft_list_peers.txt
    │   ├── consul_syslog.txt
    │   └── consul_version.txt
    ├── log
    │   └── rover.log
    ├── system
    │   ├── date.txt
    │   ├── df.txt
    │   ├── df_i.txt
    │   ├── dmesg.txt
    │   ├── hostname.txt
    │   ├── ifconfig.txt
    │   ├── last.txt
    │   ├── mount.txt
    │   ├── netstat_an.txt
    │   ├── netstat_i.txt
    │   ├── netstat_rn.txt
    │   ├── pfctl_nat.txt
    │   ├── pfctl_rules.txt
    │   ├── ps.txt
    │   ├── top.txt
    │   ├── uname.txt
    │   └── w.txt
    └── vault
        ├── vault_audit_list.txt
        ├── vault_mounts.txt
        ├── vault_status.txt
        ├── vault_syslog.txt
        └── vault_version.txt
```

The output from each command is stored in plain text files named for the command used to produce the output. `rover` also logs its own operations and stores that output in `log/rover.log`.

The next section presents a comprehensive listing of each command that is executed for a given operating system.

This is the bulk of what `rover` does:

#### Common Commands

The following system commands are run on all supported systems:

- `date`
- `df`
- `df -i`
- `df -h`
- `dmesg`
- `hostname`
- `last`
- `mount`
- `netstat -i`
- `netstat -an`
- `netstat -rn`
- `netstat -s`
- `pfctl -s rules`
- `pfctl -s nat`
- `sysctl -a`
- `uname -a`
- `w`

#### Common File Contents

- `cat /etc/fstab`
- `cat /etc/hosts`
- `cat /etc/resolv.conf`

#### Darwin Commands

The following system commands are run when Darwin is the detected system:


- `ifconfig -a`
- `ps aux`
- `top -l 1`

#### Darwin File Contents

- `/etc/fstab`
- `/etc/hosts`
- `/etc/resolv.conf`

#### FreeBSD Commands

The following system commands are run when FreeBSD is the detected system:

- `ifconfig -a`
- `iostat -dIw 1 -c 5`
- `pkg info`
- `ps aux`
- `sysctl -a`
- `top -n -b`
- `vmstat 1 10`

#### FreeBSD File Contents

- `/etc/fstab`
- `/etc/hosts`
- `/etc/resolv.conf`
- `/var/log/messages`
- `/etc/rc.conf`

#### Linux Commands

The following system commands are run when Linux is the detected system:

- `find /proc/net/bonding/ -type f -print -exec cat {} ;`
- `ls -l /dev/disk/by-id`
- `dmesg`
- `dpkg -l`
- `free -m`
- `ifconfig -a`
- `iostat -mx 1 5`
- `ip addr`
- `lsb_release`
- `ps -aux`
- `rpm -qa`
- `find /sys/class/net/ -type l -print -exec cat {}/statistics/rx_crc_errors ;`
- `find /sys/block/ -type l -print -exec cat {}/queue/scheduler ;`
- `sestatus -v`
- `swapctl -s`
- `swapon -s`
- `top -n 1 -b`
- `vmstat 1 10`

Information from distributions which use systemd:

- `journalctl --dmesg --no-pager`
- `journalctl --system", "--no-pager`
- `systemctl --all --no-pager`
- `systemctl list-unit-files --no-pager`

#### Linux File Contents

- `/etc/fstab`
- `/etc/hosts`
- `/etc/resolv.conf`
- `/var/log/daemon`
- `/var/log/debug`
- `/etc/security/limits.conf`
- `/var/log/kern.log`
- `/var/log/messages`
- `/var/log/syslog`
- `/var/log/system.log`

#### Consul Commands

- `consul version`
- `consul info`
- `consul members`
- `consul operator raft list-peers`

#### Nomad Commands

- `nomad version`
- `nomad status`
- `nomad operator raft list-peers`

#### Vault Commands

- `vault version`
- `vault audit-list`
- `vault auth -methods`
- `vault status`

## Who

### Brian Shumate [@brianshumate](https://github.com/brianshumate)

### Contributors

The fine people named in [CONTRIBUTORS.md](/blob/master/CONTRIBUTORS.md) get credit for their help on `rover` as well. Thanks everyone!
