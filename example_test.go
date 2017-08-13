package log_test

import "blitiri.com.ar/go/log"

func Example() {
	log.Init() // only needed once.

	log.Debugf("debugging information: %v %v %v", 1, 2, 3)
	log.Infof("something normal happened")
	log.Errorf("something bad happened")

	if log.V(3) { // only entered if -v was >= 3.
		//expensiveDebugging()
	}

	// Output:
}
