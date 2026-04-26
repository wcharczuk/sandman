package log

// Flag controls which types of messages are outputted.
//
// Flags should be composed with logical operations, e.g. if you want
// to enable both ERROR and WARN messages, you'd compose a flag with
//
//	myFlag := log.FlagError | log.FlagWarn
//
// The default flag level is everything except sill is enabled.
type Flag uint64

// Flag constants.
const (
	FlagDisabled Flag = 0
	FlagError    Flag = 1 << iota
	FlagWarn     Flag = 1 << iota
	FlagInfo     Flag = 1 << iota
	FlagDebug    Flag = 1 << iota
	FlagSilly    Flag = 1 << iota
)

// Flag "Yo dawg I heard you liked `Flag`"
func (f Flag) Flag() Flag {
	return f
}

type FlagProvider interface {
	Flag() Flag
}
