
# msg

A simple go logger forked from https://github.com/apsdehal/go-logger for easy logging in your programs.
Allows setting custom format for messages, has WIP JSON support, and other nice features - but keeping things light (0 external imports).

# Install

```sh
go get -v -u github.com/nexus166/msg
```

# Example

Example [program](msg/main.go) demonstrates how to use the logger. See below for __formatting__ instructions.

```go
package main

import (
	"os"

	"github.com/nexus166/msg"
)

var l *msg.Logger

func init() {
	var err error
	l, err = msg.New(msg.CLIFormat, msg.CLITimeFmt, "msg", true, os.Stdout)
	if err != nil {
		panic(err)
	}
	l.Critical("configured!")
}

func main() {
	for range []int{0, 1} {
		//l, _ = msg.New(x, os.Stderr)
		// Critically log critical
		l.Critical("This is Critical!")
		// Debug
		l.Debug("This is Debug!")
		// Give the Warning
		l.Warning("This is Warning!")
		l2, _ := msg.New(msg.JSONFormat, msg.DetailedTimeFmt, "msg", true, os.Stdout)
		// Show the error
		l2.Error("This is Error!")
		// Notice
		l2.Notice("This is Notice!")
		//Show the info
		l.Info("This is Info!")
	}
}
```

## License

The [BSD 3-Clause license](http://opensource.org/licenses/BSD-3-Clause), the same as the [Go language](http://golang.org/LICENSE) and the original project https://github.com/apsdehal/go-logger
