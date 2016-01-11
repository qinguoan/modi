package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func StrInStrings(s string, list []string) bool {
	res := false

	for _, v := range list {
		if s == v {
			res = true
		}
	}

	return res
}

func StrInIntMap(s string, d map[string]int) bool {
	if _, ok := d[s]; ok {
		return true
	}
	return false
}

func MapStrStringsCmp(m1, m2 map[string][]string) bool {
	lengthM1, lengthM2 := len(m1), len(m2)

	eq := true

	if lengthM1 != lengthM2 {

		eq = false

	} else {

		for k, v1 := range m1 {
			if len(v1) != len(m2[k]) {
				eq = false
				break
			}

			d := ListToMapCount(m2[k])

			for _, v := range v1 {
				if _, ok := d[v]; !ok {
					eq = false
					break
				}
			}

			if eq == false {
				break
			}

		}
	}
	return eq
}

func CheckLegalChar(m map[string]string) map[string]string {

	mapping := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r == '/' || r == '.' || r == '-':
			return r
		case r == ',' || r == 32:
			return -1
		default:
			return '_'
		}
	}

	alterM := make(map[string]string)

	for k, v := range m {
		alterM[k] = strings.Map(mapping, v)
	}

	return alterM
}

func TripStringsBlank(s []string) []string {

	var strs []string

	for _, value := range s {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		strs = append(strs, value)
	}
	return strs
}

func CutHttpHead(s string) string {

	re := regexp.MustCompile("^https?.//.*?/")
	s = re.ReplaceAllString(s, "/")

	return s
}

func FindConfigPath(s string, ss []string) (string, bool) {

	str := ""
	count := strings.Count(s, "/")
	changed := false
	s = CutHttpHead(s)
	for i := range ss {
		reverseValue := ss[len(ss)-1-i]
		cCount := strings.Count(reverseValue, "/")

		if cCount > count {
			continue
		} else if ok := strings.Index(s, reverseValue); (ok == 0 && cCount <= count) || reverseValue == s {
			str = reverseValue
			changed = true
			break
		}
	}

	if !changed {
		str = s
	}

	str = changeToLegalString(str)

	return str, changed
}

/*
	If > x80 found, cut behind characters and last slash with continued chars.
*/
func changeToLegalString(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			s = s[:i]
			break
		}
	}
	return s
}

func UTF8Filter(s []string) []string {
	var charPath []string

	for _, v := range s {
		v = changeToLegalString(v)
		charPath = append(charPath, v)
	}

	return charPath
}

func CurrentMilliSecond() int64 {
	current := time.Now()
	current.Year()
	curNewFormat := time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), current.Minute(), current.Second(), current.Nanosecond(), current.Location())
	return curNewFormat.UnixNano() / 1000000
}

func CurrentStamp(n int) (int64, error) {
	current := time.Now()
	min := current.Minute()
	sec := current.Second()
	sec += min * 60
	var currentFormat time.Time
	for i := 0; i <= 3600; i += n {
		if i > sec {
			currentFormat = time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), min, (i - n - min*60), 0, current.Location())
			break
		}
	}
	return currentFormat.Unix() * 1000, nil
}

func ParseFloat64(s string) (float64, error) {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func  ParseInt64(s string) (int64, error) {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
} 

func ListToMapCount(m []string) map[string]int {
	dict := make(map[string]int)
	for _, p := range m {
		if _, ok := dict[p]; ok {
			dict[p] += 1
		} else {
			dict[p] = 1
		}
	}
	return dict
}

func MapKeyToList(m map[string]int) (list []string) {

	record := make(map[string]int)

	for s, _ := range m {
		if _, ok := record[s]; !ok {
			list = append(list, s)
			record[s] = 1
		}

	}
	return
}
func TagsToKey(list []string, data map[string]string) (map[string]string, string) {

	key := ""
	tags := make(map[string]string)
	for _, v := range list {
		if value, ok := data[v]; ok && value != "" {
			key += "|" + value
			tags[v] = value
		}
	}

	return tags, key

}

func AppendListToList(src []string, dst []string) []string {

	for _, item := range src {

		dst = append(dst, item)

	}

	return dst
}
