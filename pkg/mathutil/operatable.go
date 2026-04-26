package mathutil

// Operatable are types that can be mathed.
//
// Strictly it is the [cmp.Ordered] types less string.
type Operatable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Signed are types that can be mathed with signs.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}
