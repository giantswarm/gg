package formatter

type Format string

const (
	JSON Format = "json"
	None Format = "none"
	Text Format = "text"
)

func Is(s string) Format {
	if len(s) == 0 {
		return None
	}
	if s[0] == '{' {
		return JSON
	}

	return Text
}
