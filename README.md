## Setup

go version go1.9.2

## Flags

`--debug`
Enable debug output. Optional.
`--fg, -f`
Stay in the foreground. Optional.
`--notifypid`
Send USR1 to the specified process after successful mount. 
It used internally for daemonization.

## API (Commands)

`create ORIGIN`

`delete ORIGIN`

`mount ORIGIN MOUNTPOINT`

`unmount MOUNTPOINT`

## Next Issues

