package colour

import "fmt"

const (
	format = "\033[38;5;%dm%s\033[39;49m"
)

func Blue(s string) string {
	return sprintf(67, s)
}

func DarkGreen(s string) string {
	return sprintf(72, s)
}

func LightGreen(s string) string {
	return sprintf(114, s)
}

func DarkRed(s string) string {
	return sprintf(125, s)
}

func LightRed(s string) string {
	return sprintf(161, s)
}

func sprintf(c int, s string) string {
	return fmt.Sprintf(format, c, s)
}
