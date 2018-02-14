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

Create a new filesystem. 
Now this command only checks if ORIGIN directory exists and create it if it is not exist. Also this command create config file for this filesystem (wizefs.conf).

### create Issues

* Fix it with two types (directory, zip files)
* Fix creating config file for existing ORIGIN (just check and load it)
* Check if filesystem is (isn't) mounted. Perhaps should add flag for auto-mounting after creating.

`delete ORIGIN`

Delete an existing filesystem.
Now this command only checks if ORIGIN directory exists and delete it in this case with config file.

### delete Issues

* Check if filesystem is mounted. Perhaps should add flag for auto-unmounting before deleting.

`mount ORIGIN MOUNTPOINT`

Mount an existing ORIGIN (directory or zip file) into MOUNTPOINT.
Also this command add filesystem (with all needed data) to common config (wizedb.conf).

`unmount MOUNTPOINT`

Unmount an existing MOUNTPOINT.
Also this command delete filesystem from common config.


### API Commands Issues

* Add some other Filesystems API, like `find`, `list`
* Add Files API: `load`, `get`, `remove`, `search`
* Add Internal API: `verify`


## Next Issues

* Write Bash tests, Unit tests
* Write stress tests
* Develop third filesystem type (3) to combine ZipFS and LoopbackFS ideas
* Develop filesystem design for future versions
* etc