package colour

type Palette struct {
	Key   func(s string) string
	Value func(s string) string
}

func NewNoColourPalette() Palette {
	return Palette{
		Key:   func(s string) string { return s },
		Value: func(s string) string { return s },
	}
}
