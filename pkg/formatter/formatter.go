package formatter

import (
	"fmt"
	"time"
)

type Formatter interface {
	Format(templateText string, args ...interface{}) string
	FormatTime(t time.Time, offset int) string
}

type formatter struct {
}

const TimeFormat = "02.01 15:04"

func NewFormatter() Formatter {
	return &formatter{}
}

func (f *formatter) Format(templateText string, args ...interface{}) string {
	return fmt.Sprintf(templateText, args...)
}

func (f *formatter) FormatTime(t time.Time, offset int) string {
	dur := time.Duration(offset)
	return t.Add(time.Hour * dur).Format(TimeFormat)
}
