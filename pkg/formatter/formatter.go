package formatter

type Format string

const (
	JSON Format = "json"
	None Format = "none"
	Text Format = "text"
)

func Is(l string) Format {
	if len(l) == 0 {
		return None
	}
	if l[0] == '{' {
		return JSON
	}

	return Text
}
