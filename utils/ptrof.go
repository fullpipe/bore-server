package utils

import "golang.org/x/exp/constraints"

func PtrOfBool(v bool) *bool {
	return &v
}

func PtrOfInt64(v int64) *int64 {
	return &v
}

func PtrOfInt32(v int32) *int32 {
	return &v
}

func PtrOfString(v string) *string {
	return &v
}

func PtrOfBoolType[OutT ~bool, T ~bool](in T) *OutT {
	v := OutT(in)

	return &v
}

func PtrOfStringTypePtr[OutT ~string, T ~string, PtrT *T](pt PtrT) *OutT {
	if pt == nil {
		return nil
	}

	v := OutT(*pt)

	return &v
}

func PtrOfIntegerTypePtr[OutT constraints.Integer, T constraints.Integer, PtrT *T](pt PtrT) *OutT {
	if pt == nil {
		return nil
	}

	v := OutT(*pt)

	return &v
}
