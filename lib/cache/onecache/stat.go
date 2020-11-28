package onecache

import "time"

type stat struct {
	totalOperations uint32

	totalHits   uint32
	totalMisses uint32

	totalReads  uint32
	totalWrites uint32

	timeStart int64

	totalReadBytes  int64
	totalWriteBytes int64
}

func (_this *stat) reset() {
	_this.totalOperations = 0
	_this.totalHits = 0
	_this.totalMisses = 0
	_this.totalReads = 0
	_this.totalWrites = 0
	_this.timeStart = time.Now().Unix()
	_this.totalReadBytes = 0
	_this.totalWriteBytes = 0
}