package formatter

import "fmt"

type Formatter interface {
	Format(templateText string, args ...interface{}) string
}

type formatter struct {
}

func NewFormatter() Formatter {
	return &formatter{}
}

func (f *formatter) Format(templateText string, args ...interface{}) string {
	return fmt.Sprintf(templateText, args...)
}
