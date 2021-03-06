// Code generated by "stringer -type=Retryable -trimprefix=Retryable"; DO NOT EDIT.

package request

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RetryableUnknown-0]
	_ = x[RetryableYes-1]
	_ = x[RetryableNo-2]
}

const _Retryable_name = "UnknownYesNo"

var _Retryable_index = [...]uint8{0, 7, 10, 12}

func (i Retryable) String() string {
	if i < 0 || i >= Retryable(len(_Retryable_index)-1) {
		return "Retryable(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Retryable_name[_Retryable_index[i]:_Retryable_index[i+1]]
}
