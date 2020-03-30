[![GoDoc](https://godoc.org/github.com/giantswarm/gg?status.svg)](http://godoc.org/github.com/giantswarm/gg)
[![CircleCI](https://circleci.com/gh/giantswarm/gg.svg?style=shield)](https://circleci.com/gh/giantswarm/gg)

# gg

A simple JSON logs parser for Kubernetes Operators designed for effective
debugging.



## Usage

```
A simple JSON logs parser for Kubernetes Operators designed for effective debugging.

Usage:
  gg [flags]
  gg [command]

Examples:
    The following examples make use of a basic.json file which contains JSON
    logs where each log line is a single JSON object. You can find the
    basic.json file in ./cmd/fixture along with other test fixtures used for
    golden file tests.

    Select all logs where any key matches "obj" and its associated value matches
    "qihx8". This can be used to e.g. grep for all logs related to Tenant
    Cluster "qihx8".

        cat basic.json | gg -s obj:qihx8

    Select all logs like the example above but on top of that filter also for
    logs where any key matches "res" and its associated value matches "dra".
    This can be used to e.g. grep for all logs of the "drainer" and
    "drainfinisher" resource implementation.

        cat basic.json | gg -s obj:qihx8 -s res:dra

    Select all logs like the example above but on top of that only output
    key-value pairs of the logs where any key matches "ti" or "mes". This can be
    used to e.g. show only "time" and "message". Note that the order of fields
    given determines the output order. Here "ti,mes" makes it way easier to read
    the output since "time" is always consistently formatted, whereas "message"
    can be of almost arbitrary length.

        cat basic.json | gg -s obj:qihx8 -s res:dra -f ti,mes

    Select all logs like the example above but on top of that group output
    key-value pairs of the logs based on the common value of associated keys
    matching "lo". This can be used to e.g. group resource logs by their
    reconciliation "loop".

        cat basic.json | gg -s obj:qihx8 -s res:dra -f ti,mes -g lo

    Display the list of resources executed for a given CR.

        cat basic.json | gg -s obj:qihx8 -s con:mac -f res -o text -g lo | uniq

Available Commands:
  help        Help about any command
  version     Print version information.

Flags:
  -f, --field strings    Fields the output lines should contain only.
  -g, --group string     Group logs by inserting an empty line after the group end.
  -h, --help             help for gg
  -o, --output string    Output format, either json or text. (default "json")
  -s, --select strings   Select lines based on the given key:val regular expression.

Use "gg [command] --help" for more information about a command.
```
