# Formatting

By default all log messages have format that you can see above (on pic).
But you can override the default format and set format that you want.

You can do it for Logger instance (after creating logger) ...
```go
log, _ := logger.New("pkgname", 1)
log.SetFormat(format)
```
... or for package
```go
logger.SetDefaultFormat(format)
```
If you do it for package, all existing loggers will print log messages with format that these used already.
But all newest loggers (which will be created after changing format for package) will use your specified format.

But anyway after this, you can still set format of message for specific Logger instance.

Format of log message must contains verbs that represent some info about current log entry.
Ofc, format can contain not only verbs but also something else (for example text, digits, symbols, etc)

### Format verbs:
You can use the following verbs:
```
%{id}           - means number of current log message
%{module}       - means module name (that you passed to func New())
%{time}			- means current time in format "2006-01-02 15:04:05"
%{time:format}	- means current time in format that you want
					(supports all formats supported by go package "time")
%{level}		- means level name (upper case) of log message ("ERROR", "DEBUG", etc)
%{lvl}			- means first 3 letters of level name (upper case) of log message ("ERR", "DEB", etc)
%{file}			- means name of file in what you wanna write log
%{filename}		- means the same as %{file}
%{line}			- means line number of file in what you wanna write log
%{message}		- means your log message
```
Non-existent verbs (like ```%{nonex-verb}``` or ```%{}```) will be replaced by an empty string.
Invalid verbs (like ```%{inv-verb```) will be treated as plain text.
