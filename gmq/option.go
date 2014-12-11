package gmq

import (
	"errors"
	"time"
)

// Tried to use option type for nullable columns, but found that may be a problem for json encoding
// Will not use this for now

var (
	ErrOptionIsNone = errors.New("Option is not valid.")
)

type Option interface {
	IsDefined() bool
}

type OptionInt interface {
	Option
	Get() (int, error)
}

func SomeInt(n int) OptionInt { return _OptionInt{n, true} }
func NoneInt() OptionInt      { return _OptionInt{0, false} }

type OptionInt64 interface {
	Option
	Get() (int64, error)
}

func SomeInt64(n int64) OptionInt64 { return _OptionInt64{n, true} }
func NoneInt64() OptionInt64        { return _OptionInt64{0, false} }

type OptionString interface {
	Option
	Get() (string, error)
}

func SomeString(n string) OptionString { return _OptionString{n, true} }
func NoneString() OptionString         { return _OptionString{"", false} }

type OptionTime interface {
	Option
	Get() (time.Time, error)
}

func SomeTime(n time.Time) OptionTime { return _OptionTime{n, true} }
func NoneTime() OptionTime            { return _OptionTime{time.Now(), false} }

type OptionFloat64 interface {
	Option
	Get() (float64, error)
}

func SomeFloat64(n float64) OptionFloat64 { return _OptionFloat64{n, true} }
func NoneFloat64() OptionFloat64          { return _OptionFloat64{0, false} }

type _OptionInt struct {
	value int
	valid bool
}

func (o _OptionInt) Get() (int, error) {
	if o.IsDefined() {
		return o.value, nil
	} else {
		return 0, ErrOptionIsNone
	}
}

func (o _OptionInt) IsDefined() bool { return o.valid }

type _OptionInt64 struct {
	value int64
	valid bool
}

func (o _OptionInt64) Get() (int64, error) {
	if o.IsDefined() {
		return o.value, nil
	} else {
		return 0, ErrOptionIsNone
	}
}

func (o _OptionInt64) IsDefined() bool { return o.valid }

type _OptionFloat64 struct {
	value float64
	valid bool
}

func (o _OptionFloat64) Get() (float64, error) {
	if o.IsDefined() {
		return o.value, nil
	} else {
		return 0, ErrOptionIsNone
	}
}

func (o _OptionFloat64) IsDefined() bool { return o.valid }

type _OptionString struct {
	value string
	valid bool
}

func (o _OptionString) Get() (string, error) {
	if o.IsDefined() {
		return o.value, nil
	} else {
		return "", ErrOptionIsNone
	}
}

func (o _OptionString) IsDefined() bool { return o.valid }

type _OptionTime struct {
	value time.Time
	valid bool
}

func (o _OptionTime) Get() (time.Time, error) {
	if o.IsDefined() {
		return o.value, nil
	} else {
		return time.Now(), ErrOptionIsNone
	}
}

func (o _OptionTime) IsDefined() bool { return o.valid }
