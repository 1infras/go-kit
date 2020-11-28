package consume

import "time"

type stat struct {
	totalOperations uint32
	totalConsumed  uint32
	totalDispeared  uint32
	totalRetry      uint32
	totalErrors     uint32
	timeStart       int64

	totalReceivedBytes int64
}

func (_this *stat) reset() {
	_this.totalOperations = 0
	_this.totalConsumed = 0
	_this.totalDispeared = 0
	_this.totalRetry = 0
	_this.totalErrors = 0
	_this.timeStart = time.Now().Unix()
	_this.totalReceivedBytes = 0
}
