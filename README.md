# cirsim
Simple spice-based simulator

# Compile and Run

1. Pull repository to `$GOPATH/src/github.com/felipeek/`
2. Get `go-chart` by running `go get -u github.com/wcharczuk/go-chart`
3. Cd to `$GOPATH/src/github.com/felipeek/cirsim` and run `compile.bat` on Windows or `compile.sh` on Linux.
4. The binary will be available on `$GOPATH/bin/cirsim`
5. To run you must provide a spice file and, if you are doing a transient analysis, you must ask to generate the result graphs.

```
cirsim parameters:
-graphs
   Generate graphs
-path string
   Spice file path`
```

For example,

`cirsim -path res/custom.sp -graphs`
