# rover
                                                                  _.=
    .----.-----.--.--.-----.----.                            +=====|
    |   _|  _  |  |  |  -__|   _|          ----========      |-___-|
    |__| |_____|\___/|_____|__|    ----================   @--|@---@|--@

## What?

`rover` is a tool that explores systems and reports on what it finds.

It is inspired by tools like my friend Brent's [debug-ninja](https://github.com/fprimex/debug-ninja) and Apple's `sysdiagnose`.

`rover` gathers important factoids from a system to paint a detailed picture of the current operating environment for engineers and operators curious about such things.

The types of information `rover` gathers includes:

- Operating system command output
- Operating system logging
- Application or service specifc command output
- Application or service specific logging

All of the stored information can then be packaged up into a zip file named for the host, and shared however you prefer. Soon, `rover` will also support shipping the zip file to destinations like an S3 bucket.

## Why?

To assist with troubleshooting of systems and help make the process efficient, reliable, and repeatable.

## How?

A CLI tool called `rover` that is written in [Go](https://golang.org/)

## Building

If you have a Go environment established, you can install and run the
`rover` command like this:

```
$ go get github.com/brianshumate/rover
$ rover
Usage: rover [--version] [--help] <command> [<args>]

Available commands are:
    archive    Archive rover data into zip file
    consul     Execute Consul related commands and store output
    info       Output installation information
    nomad      Execute Nomad related commands and store output
    system     Execute system commands and store output
    upload     Uploads rover archive file to S3 bucket
    vault      Execute Vault related commands and store output
```

Otherwise, you can consult the [Go documentation for the Go tools](https://golang.org/doc/install#install) for you particular platform, and _go_ from there!

If you visit `$GOPATH/src/github/brianshumate/rover` you can build releases of the `rover` binary for different platforms by typing `make`.

The project currently builds binaries for FreeBSD, Linux, and macOS by default; they're located in the following subdirectories of `pkg/` after a successful build:

```
├── pkg
│   ├── darwin-amd64
│   │   └── rover
│   ├── freebsd-amd64
│   │   └── rover
│   └── linux-amd64
│       └── rover
```

### Development Build

If you'd prefer to make a development build for just your own host architecture and OS, you can use `make dev` and the `rover` binary will also be located in `./bin` after successful compiliation.

## Running

It's a single binary, `rover` with built in help:

```
$ make dev
...
$ ./bin/rover
Usage: rover [--version] [--help] <command> [<args>]

Available commands are:
    archive    Archive rover data as zip file
    consul     Execute Consul related commands and store their output
    nomad      Execute Nomad related commands and store their output
    system     Execute system commands and store their output
    vault      Execute Vault related commands and store their output
```

###  Environment Variables

Currently all configuration is specified as flags or set as environment variables. For detailed help, including available flags, use `rover <command> --help`.

Here are the environment variables:

- `AWS_ACCESS_KEY_ID`: Access key ID for AWS
- `AWS_SECRET_ACCESS_KEY`: Secret access key ID for AWS
- `AWS_BUCKET`: Name of the S3 bucket
- `AWS_PREFIX`: Filename prefix
- `AWS_REGION`: AWS region for the bucket

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

#### Vault Commands

- `vault version`
- `vault audit-list`
- `vault auth -methods`
- `vault status`

#### Command Combinations

You can chain commands together to build a zip file with your desired contents like this:

```
./darwin-amd64/rover consul && \
./darwin-amd64/rover vault && \
./darwin-amd64/rover system && \
./darwin-amd64/rover archive
```

Investigation into simplified meta commands and easy one-liners is also on the roadmap.

## Who

### Brian Shumate

[@brianshumate](https://github.com/brianshumate)

### Contributors

The fine people named in [CONTRIBUTORS.md](/blob/master/CONTRIBUTORS.md) get credit for their help on `rover` as well. Thanks everyone!
