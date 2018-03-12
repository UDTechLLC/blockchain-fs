## Setup

go version go1.9.2

System: Linux amd64.

## Install

GUI application is based on platform-native GUI library `andlabs/ui`, we have fork of this library `leedark/ui` and you should get it by `go get`:

```
go get -u github.com/leedark/ui
```

Then you should go to the directory `wizefs\ui` and run `go build`.

## Main window

![main-window](images/main-window.png)



## Create filesystem dialog

![create-dialog](images/create-dialog.png)



## Mount filesystem

![main-window-unmounted](images/main-window-unmounted.png)



## Unmount filesystem, put and get file

![main-window-mounted](images/main-window-mounted.png)