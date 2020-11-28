package onecache

import (
	"strconv"
	"time"
)

const (
	AddElement int = iota
	DeleteElement
)

type Element interface {
	Bytes() []byte
	Val() interface{}
	Int() (int, error)
	Int64() (int64, error)
	Float64() (float64, error)
	Boolean() (bool, error)
	String() string
}

type item struct {
	key        string
	value      interface{}
	expiration time.Duration
	action     int
}

type element struct {
	value []byte
}

func (_this *element) Bytes() []byte {
	return _this.value
}

func (_this *element) Val() interface{} {
	return _this.value
}

func (_this *element) Int() (int, error) {
	return strconv.Atoi(string(_this.value))
}

func (_this *element) Int64() (int64, error) {
	return strconv.ParseInt(string(_this.value), 10, 64)
}

func (_this *element) Float64() (float64, error) {
	return strconv.ParseFloat(string(_this.value), 64)
}

func (_this *element) Boolean() (bool, error) {
	return strconv.ParseBool(string(_this.value))
}

func (_this *element) String() string {
	return string(_this.value)
}
