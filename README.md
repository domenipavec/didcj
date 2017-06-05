# didcj
A tool for running distributed codejam code

## didcj local

Run dcj locally using *dcj.sh*

Example:
`didcj local --nodes 10`

## didcj remote

Run dcj on remote

First start the nodes:
`didcj remote start --nodes 100`

Optionally only start the daemon on running nodes:
`didcj remote start --nodes 100 --daemon`

Run:
`didcj remote --nodes 100`

At the end stop the nodes:
`didcj remote stop`

## didcj generate

### didcj generate config

Generate typical config file

### didcj generate main <filename>

Generate base code file with filename.

### didcj generate input

Generate input header file based on config

Input function generators: (duration_ns has actual resolution of ~500ns)
- CONSTANT: constant `value`
- RANDOM_RANGE: random between `min` and `max`
- INCREASING_RANDOM_RANGE: random in window increasingly divided between `min` and `max`
- DECREASING_RANDOM_RANGE: random in window decreasingly divided between `min` and `max`
- RANDOM_LIST: random between values in array `values`
