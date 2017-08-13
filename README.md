
# blitiri.com.ar/go/log

[![GoDoc](https://godoc.org/blitiri.com.ar/go/log?status.svg)](https://godoc.org/blitiri.com.ar/go/log)

[log](https://godoc.org/blitiri.com.ar/go/log) is a Go package implementing a
simple logger.

It implements an API somewhat similar to [glog](github.com/google/glog), with
a focus towards simplicity and integration with standard tools such as
systemd.


## Examples

```go
log.Init()  // only needed once.

log.Debugf("debugging information: %v", x)
log.Infof("something normal happened")
log.Errorf("something bad happened")
log.Fatalf("tragic")

if log.V(3) {  // only entered if -v was >= 3.
	expensiveDebugging()
}
```


## Status

The API should be considered stable.

Branch v1 will only have backwards-compatible changes made to it.
There are no plans for v2 at the moment.


## Contact

If you have any questions, comments or patches please send them to
albertito@blitiri.com.ar.

