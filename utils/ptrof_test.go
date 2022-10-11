package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtrOfStringTypePtr_ConvertsNilToNil(t *testing.T) {
	type A string
	type B string

	var a *A
	b := B("aaa")

	result := PtrOfStringTypePtr[B](a)
	assert.IsType(t, &b, result)
	assert.Nil(t, result)
}

func TestPtrOfStringTypePtr_ConvertsEmptyToEmpty(t *testing.T) {
	type A string
	type B string

	var a A
	var b B

	result := PtrOfStringTypePtr[B](&a)
	assert.IsType(t, &b, result)
	assert.NotNil(t, result)
	assert.Equal(t, &b, result)
}

func TestPtrOfStringTypePtr_ConvertsPtrToStringPtr(t *testing.T) {
	type A string

	a := A("aaa")
	b := "aaa"

	result := PtrOfStringTypePtr[string](&a)
	assert.IsType(t, &b, result)
	assert.NotNil(t, result)
	assert.Equal(t, &b, result)
}

func TestPtrOfStringTypePtr_ConvertsPtrToPtr(t *testing.T) {
	type A string
	type B string

	a := A("aaa")
	b := B("aaa")

	result := PtrOfStringTypePtr[B](&a)
	assert.IsType(t, &b, result)
	assert.NotNil(t, result)
	assert.Equal(t, &b, result)
}

func TestPtrOfIntegerTypePtr_ConvertsPtrToPtr(t *testing.T) {
	type A int64
	type B int64

	a := A(1)
	b := B(1)

	result := PtrOfIntegerTypePtr[B](&a)
	assert.IsType(t, &b, result)
	assert.NotNil(t, result)
	assert.Equal(t, &b, result)
}

func TestPtrOfBoolType_ConvertsBoolToPtr(t *testing.T) {
	type B bool

	a := true
	b := B(true)

	result := PtrOfBoolType[B](a)
	assert.IsType(t, &b, result)
	assert.NotNil(t, result)
	assert.Equal(t, &b, result)
}
