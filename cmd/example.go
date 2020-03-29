package cmd

var example = `    The following examples make use of a test.json file which contains JSON logs
    where each log line is a single JSON object.

    Grep for all logs where any key matches "obj" and its associated value
    matches "qihx8". This can be used to e.g. grep for all logs related to
    Tenant Cluster "qihx8".

        cat test.json | gg -g obj:qihx8

    Grep for all logs like the example above but on top of that filter also for
    logs where any key matches "res" and its associated value matches "dra".
    This can be used to e.g. grep for all logs of the "drainer" and
    "drainfinisher" resource implementation.

        cat test.json | gg -g obj:qihx8 -g res:dra

    Grep for all logs like the example above but on top of that only output
    key-value pairs of the logs where any key matches "ti" or "mes". This can be
    used to e.g. show only "time" and "message". Note that the order of fields
    given determines the output order. Here "ti,mes" makes it way easier to read
    the output since "time" is always consistently formatted, whereas "message"
    can be of almost arbitrary length.

        cat test.json | gg -g obj:qihx8 -g res:dra -f ti,mes`
