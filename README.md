
# blitiri.com.ar/go/log

[![GoDoc](https://godoc.org/blitiri.com.ar/go/log?status.svg)](https://godoc.org/blitiri.com.ar/go/log)
[![Build Status](https://travis-ci.org/albertito/log.svg?branch=master)](https://travis-ci.org/albertito/log)
[![Go Report Card](https://goreportcard.com/badge/github.com/albertito/log)](https://goreportcard.com/report/github.com/albertito/log)

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

The API should be considered generally stable, and no backwards-incompatible
changes are expected.

Some specific symbols are experimental, and are marked as such in their
documentation.  Those might see backwards-incompatible changes, including
removing them entirely.


## Contact

If you have any questions, comments or patches please send them to
albertito@blitiri.com.ar.

