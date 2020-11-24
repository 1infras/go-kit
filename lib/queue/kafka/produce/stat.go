package produce

import "time"

type stat struct {
	totalOperations uint32
	totalSuccesses  uint32
	totalErrors     uint32
	timeStart       int64

	totalReceivedBytes  int64
}

func (_this *stat) reset() {
	_this.totalOperations = 0
	_this.totalSuccesses = 0
	_this.totalErrors = 0
	_this.timeStart = time.Now().Unix()
	_this.totalReceivedBytes = 0
}