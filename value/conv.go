package value

import (
	"bytes"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mithrandie/ternary"
)

type DatetimeFormatMap map[string]string

func (m DatetimeFormatMap) Get(s string) string {
	if f, ok := m[s]; ok {
		return f
	}
	f := ConvertDatetimeFormat(s)
	m[s] = f
	return f
}

var DatetimeFormats = DatetimeFormatMap{}

func StrToTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	location, err := time.LoadLocation(s)
	if err != nil {
		return time.Now(), err
	}

	if 0 < len("") {
		if t, e := time.ParseInLocation(DatetimeFormats.Get(""), s, location); e == nil {
			return t, nil
		}
	}

	if 8 <= len(s) && '0' <= s[0] && s[0] <= '9' {
		switch {
		case s[4] == '-':
			if len(s) < 10 {
				if t, e := time.ParseInLocation("2006-1-2", s, location); e == nil {
					return t, nil
				}
			} else if len(s) == 10 {
				if t, e := time.ParseInLocation("2006-01-02", s, location); e == nil {
					return t, nil
				}
			} else if s[10] == 'T' {
				if s[len(s)-3] == '+' || s[len(s)-6] == '-' || s[len(s)-1] == 'Z' {
					if t, e := time.Parse(time.RFC3339Nano, s); e == nil {
						return t, nil
					}
				} else {
					if t, e := time.ParseInLocation("2006-01-02T15:04:05.999999999", s, location); e == nil {
						return t, nil
					}
				}
			} else if s[10] == ' ' {
				if t, e := time.ParseInLocation("2006-01-02 15:04:05.999999999", s, location); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-01-02 15:04:05.999999999 -07:00", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-01-02 15:04:05.999999999 -0700", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-01-02 15:04:05.999999999 MST", s); e == nil {
					return t, nil
				}
			} else {
				if t, e := time.ParseInLocation("2006-1-2 15:04:05.999999999", s, location); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-1-2 15:04:05.999999999 -07:00", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-1-2 15:04:05.999999999 -0700", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006-1-2 15:04:05.999999999 MST", s); e == nil {
					return t, nil
				}
			}
		case s[4] == '/':
			if len(s) < 10 {
				if t, e := time.ParseInLocation("2006/1/2", s, location); e == nil {
					return t, nil
				}
			} else if len(s) == 10 {
				if t, e := time.ParseInLocation("2006/01/02", s, location); e == nil {
					return t, nil
				}
			} else if s[10] == ' ' {
				if t, e := time.ParseInLocation("2006/01/02 15:04:05.999999999", s, location); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/01/02 15:04:05.999999999 Z07:00", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/01/02 15:04:05.999999999 -0700", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/01/02 15:04:05.999999999 MST", s); e == nil {
					return t, nil
				}
			} else {
				if t, e := time.ParseInLocation("2006/1/2 15:04:05.999999999", s, location); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/1/2 15:04:05.999999999 Z07:00", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/1/2 15:04:05.999999999 -0700", s); e == nil {
					return t, nil
				} else if t, e := time.Parse("2006/1/2 15:04:05.999999999 MST", s); e == nil {
					return t, nil
				}
			}
		default:
			if t, e := time.Parse(time.RFC822, s); e == nil {
				return t, nil
			} else if t, e := time.Parse(time.RFC822Z, s); e == nil {
				return t, nil
			}
		}
	}
	return time.Time{}, errors.New("conversion failed")
}

func ConvertDatetimeFormat(format string) string {
	runes := []rune(format)
	var buf bytes.Buffer

	escaped := false
	for _, r := range runes {
		if !escaped {
			switch r {
			case '%':
				escaped = true
			default:
				buf.WriteRune(r)
			}
			continue
		}

		switch r {
		case 'a':
			buf.WriteString("Mon")
		case 'b':
			buf.WriteString("Jan")
		case 'c':
			buf.WriteString("1")
		case 'd':
			buf.WriteString("02")
		case 'E':
			buf.WriteString("_2")
		case 'e':
			buf.WriteString("2")
		case 'F':
			buf.WriteString(".999999")
		case 'f':
			buf.WriteString(".000000")
		case 'H':
			buf.WriteString("15")
		case 'h':
			buf.WriteString("03")
		case 'i':
			buf.WriteString("04")
		case 'l':
			buf.WriteString("3")
		case 'M':
			buf.WriteString("January")
		case 'm':
			buf.WriteString("01")
		case 'N':
			buf.WriteString(".999999999")
		case 'n':
			buf.WriteString(".000000000")
		case 'p':
			buf.WriteString("PM")
		case 'r':
			buf.WriteString("03:04:05 PM")
		case 's':
			buf.WriteString("05")
		case 'T':
			buf.WriteString("15:04:05")
		case 'W':
			buf.WriteString("Monday")
		case 'Y':
			buf.WriteString("2006")
		case 'y':
			buf.WriteString("06")
		case 'Z':
			buf.WriteString("Z07:00")
		case 'z':
			buf.WriteString("MST")
		default:
			buf.WriteRune(r)
		}
		escaped = false
	}

	return buf.String()
}

func Float64ToTime(f float64) time.Time {
	s := Float64ToStr(f)
	ns := strings.Split(s, ".")
	sec, _ := strconv.ParseInt(ns[0], 10, 64)
	var nsec int64
	if 1 < len(ns) {
		if 9 < len(ns[1]) {
			ns[1] = ns[1][:9]
		}
		nsec, _ = strconv.ParseInt(ns[1]+strings.Repeat("0", 9-len(ns[1])), 10, 64)
	}
	return time.Unix(sec, nsec)
}

func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ParseFloat64(f float64) Primary {
	if math.Remainder(f, 1) == 0 {
		return NewInteger(int64(f))
	}
	return NewFloat(f)
}

func ToInteger(p Primary) Primary {
	switch p.(type) {
	case Integer:
		return p
	case Float:
		f := p.(Float).Raw()
		if math.Remainder(f, 1) == 0 {
			return NewInteger(int64(f))
		}
	case String:
		s := strings.TrimSpace(p.(String).Raw())
		if maybeNumber(s) {
			if i, e := strconv.ParseInt(s, 10, 64); e == nil {
				return NewInteger(i)
			}
			if f, e := strconv.ParseFloat(s, 64); e == nil {
				if math.Remainder(f, 1) == 0 {
					return NewInteger(int64(f))
				}
			}
		}
	}

	return NewNull()
}

func ToFloat(p Primary) Primary {
	switch p.(type) {
	case Integer:
		return NewFloat(float64(p.(Integer).Raw()))
	case Float:
		return p
	case String:
		s := strings.TrimSpace(p.(String).Raw())
		if maybeNumber(s) {
			if f, e := strconv.ParseFloat(p.(String).Raw(), 64); e == nil {
				return NewFloat(f)
			}
		}
	}

	return NewNull()
}

func maybeNumber(s string) bool {
	slen := len(s)
	if 1 < slen && (s[0] == '-' || s[0] == '+') && '0' <= s[1] && s[1] <= '9' {
		return true
	}
	if 0 < slen && '0' <= s[0] && s[0] <= '9' {
		if 8 <= slen {
			if s[4] == '-' && (s[6] == '-' || s[7] == '-') {
				return false
			}
			if s[4] == '/' && (s[6] == '/' || s[7] == '/') {
				return false
			}
			if s[2] == ' ' {
				return false
			}
		}
		return true
	}
	return false
}

func ToDatetime(p Primary) Primary {
	switch p.(type) {
	case Integer:
		dt := time.Unix(p.(Integer).Raw(), 0)
		return NewDatetime(dt)
	case Float:
		dt := Float64ToTime(p.(Float).Raw())
		return NewDatetime(dt)
	case Datetime:
		return p
	case String:
		s := strings.TrimSpace(p.(String).Raw())
		if dt, e := StrToTime(s); e == nil {
			return NewDatetime(dt)
		}
		if maybeNumber(s) {
			if i, e := strconv.ParseInt(s, 10, 64); e == nil {
				dt := time.Unix(i, 0)
				return NewDatetime(dt)
			}
			if f, e := strconv.ParseFloat(s, 64); e == nil {
				dt := Float64ToTime(f)
				return NewDatetime(dt)
			}
		}
	}

	return NewNull()
}

func ToBoolean(p Primary) Primary {
	switch p.(type) {
	case Boolean:
		return p
	case String, Integer, Float, Ternary:
		if p.Ternary() != ternary.UNKNOWN {
			return NewBoolean(p.Ternary().ParseBool())
		}
	}
	return NewNull()
}

func ToString(p Primary) Primary {
	switch p.(type) {
	case String:
		return p
	case Integer:
		return NewString(Int64ToStr(p.(Integer).Raw()))
	case Float:
		return NewString(Float64ToStr(p.(Float).Raw()))
	}
	return NewNull()
}
