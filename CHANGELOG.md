## v0.2.0

- Complete switchover to go-hclog
- More idiomatic use of CLI / UI package
- Return instead of exit where possible
- All errors and log lines updated
- gofmt all the things

## v0.1.0

- Using go-hclog for the log output
- Moved internal helpers package into command package
- Add spinner to long lived commands

## v0.0.5

- Archive now accepts these flags
  - `-keep-source` preserves the archive source directory
  - `-path` full archive file output path
- Fix typos
- Add Go Report status

## v0.0.4

- Additional info command output
- More help output touches
- Update documentation

## v0.0.3

- Initial upload command
- Improved help output
- Update documentation

## v0.0.2

- Initial info command
- Update documentation

## v0.0.1

- Initial release has some basic funcitonality
  - System command gathering
  - Consul command/profiling/syslog gathering
  - Vault commands/syslog gathering
  - Zip archive of all output
