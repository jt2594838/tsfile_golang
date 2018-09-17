package operator

import "strings"

type IntNeqFilter struct {
	ref int32
}

func (f *IntNeqFilter) satisfy(val interface{}) bool {
	if v, ok := val.(int32); ok {
		return f.ref != v
	}
	return false
}

type LongNeqFilter struct {
	ref int64
}

func (f *LongNeqFilter) satisfy(val interface{}) bool {
	if v, ok := val.(int64); ok {
		return f.ref != v
	}
	return false
}

type StrNeqFilter struct {
	ref string
}

func (f *StrNeqFilter) satisfy(val interface{}) bool {
	if v, ok := val.(string); ok {
		return strings.Compare(v, f.ref) != 0
	}
	return false
}

type FloatNeqFilter struct {
	ref float32
}

func (f *FloatNeqFilter) satisfy(val interface{}) bool {
	if v, ok := val.(float32); ok {
		return v != f.ref
	}
	return false
}

type DoubleNeqFilter struct {
	ref float64
}

func (f *DoubleNeqFilter) satisfy(val interface{}) bool {
	if v, ok := val.(float64); ok {
		return v != f.ref
	}
	return false
}
