package cmd

var example = `    The following examples make use of a test.json file which contains JSON logs
    where each log line is a single JSON object.

    Select all logs where any key matches "obj" and its associated value matches
    "qihx8". This can be used to e.g. grep for all logs related to Tenant
    Cluster "qihx8".

        cat test.json | gg -s obj:qihx8

    Select all logs like the example above but on top of that filter also for
    logs where any key matches "res" and its associated value matches "dra".
    This can be used to e.g. grep for all logs of the "drainer" and
    "drainfinisher" resource implementation.

        cat test.json | gg -s obj:qihx8 -s res:dra

    Select all logs like the example above but on top of that only output
    key-value pairs of the logs where any key matches "ti" or "mes". This can be
    used to e.g. show only "time" and "message". Note that the order of fields
    given determines the output order. Here "ti,mes" makes it way easier to read
    the output since "time" is always consistently formatted, whereas "message"
    can be of almost arbitrary length.

        cat test.json | gg -s obj:qihx8 -s res:dra -f ti,mes

    Select all logs like the example above but on top of that group output
    key-value pairs of the logs based on the common value of associated keys
    matching "lo". This can be used to e.g. group resource logs by their
    reconciliation "loop".

        cat test.json | gg -s obj:qihx8 -s res:dra -f ti,mes -g lo

    Display the list of resources executed for a given CR.

        cat test.json | gg -s obj:qihx8 -s con:mac -f res -o text -g lo | uniq`
