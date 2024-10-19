package mathutil

// Operatable are types that can be mathed.
//
// Strictly it is the `Ordered` types less string.
type Operatable interface {
	~int | ~int64 | ~float64
}
