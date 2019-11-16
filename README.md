
# msg

## Overview

msg is a Golang logging library forked from https://github.com/apsdehal/go-logger.
It has the following features I could not find anywhere else (hence the fork):
- Ability to create new or reconfigure existing log levels
- Log-level based colored output
- Custom format for messages
- Custom time format for messages
- Multiple pre-configured formats to pick from: 
  - cli
  - plain
  - plain-emoji
  - std
  - std-emoji
  - simple
  - JSON  (make sure you properly escape your messages and disable colored output!)
  - All the default time format constants (https://golang.org/pkg/time/#pkg-constants)
- **0 external imports.**


## Installation

```sh
go get -v -u github.com/nexus166/msg
```

## Usage

Example below demonstrates some ways to start using msg

```go
package main

import (
	"os"
	"strings"
	"time"

	"github.com/nexus166/msg"
)

var l *msg.Logger

func init() {
	msg.Info("You can immediately use the default Logger, or configure new Loggers")
	var err error
	// Formats is a map[string]Format, it is easy to access the various formats with a simple commandline flag
	l, err = msg.New(msg.Formats["cli"].String(), msg.Formats["rfc822"].String(), "msg-cli", true, os.Stdout, msg.LDebug) //configure the global l *msg.Logger
	if err != nil {
		msg.Fatal(err.Error())
	}
	msg.Info("New Logger configured!")
}

func main() {
	for f, F := range msg.Formats { // lets test them all
		if !strings.Contains(f, "rfc") {
			for _, b := range []bool{false, true} {
				// to use a preconfigured format, you need to pass the format string to msg.New(). For this, String() method is available on preconfigured Formats.
				l, errL := msg.New(F.String(), msg.Formats[msg.DetailedTimeFmt].String(), "msg-"+f, b, os.Stderr, msg.LDebug)
				if errL != nil {
					msg.Error(errL.Error())
				}
				for _, lvl := range msg.Levels {
					l.Log(msg.Lvl(lvl.ID), "Log "+lvl.Str+", format: "+f)
				}
			}
		}
	}
	for _, lvl := range msg.Levels {
		// send a message in this project's original format. You can pass constants directly from time lib to msg as a time format.
		l, _ = msg.New("#%[1]d %[2]s %[4]s:%[5]d ▶ %.3[6]s %[7]s", time.Kitchen, "custom-formats", true, os.Stdout, msg.LDebug)
		l.Log(msg.Lvl(lvl.ID), "Log "+lvl.Str+" legacy format")
	}
}
```

[Running the above code](https://play.golang.org/p/srOvuLkIvWe) will demo all available formats (and the customized one):

![image](https://user-images.githubusercontent.com/9354925/68991311-c519bb00-085d-11ea-8e00-98853feeec09.png)


## License

The [BSD 3-Clause license](http://opensource.org/licenses/BSD-3-Clause), the same as the [Go language](http://golang.org/LICENSE) and the original project https://github.com/apsdehal/go-logger
