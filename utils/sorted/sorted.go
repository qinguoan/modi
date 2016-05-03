/*
	This package implement sort for list and dict, using specified sort function.

	Author: qinguoan@wandoujia.com 2015-08-16
*/

package sorted

import (
	"reflect"
	"strings"
)

/*
	This part is used to sort specified string list by length number of each string.
*/

type ByStringLen []string

func (s ByStringLen) Len() int {
	return len(s)
}

func (s ByStringLen) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByStringLen) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

/*
	This part is used to sort specified string list by key word count of each string.
*/

type CharCount struct {
	List []string
	Sep  string
}

func ByCharCount(list []string, sep string) CharCount {
	cc := CharCount{
		List: list,
		Sep:  sep,
	}

	return cc

}

func (s CharCount) Len() int {
	return len(s.List)
}

func (s CharCount) Swap(i, j int) {
	s.List[i], s.List[j] = s.List[j], s.List[i]
}

func (s CharCount) Less(i, j int) bool {
	return strings.Count(s.List[i], s.Sep) < strings.Count(s.List[j], s.Sep)
}

/*
	This part is used to sort specified map by value of map.
*/

type MapValue struct {
	Dict map[string]interface{}
	List []string
	Kind reflect.Kind
	Succ bool
}

func ByMapValue(m map[string]interface{}) MapValue {
	mapValue := MapValue{
		Dict: m,
		Succ: true,
	}
	for k, v := range mapValue.Dict {
		if mapValue.Kind == reflect.Invalid {
			mapValue.Kind = reflect.TypeOf(v).Kind()
		} else if mapValue.Kind != reflect.TypeOf(v).Kind() {
			mapValue.Succ = false
			break
		}
		mapValue.List = append(mapValue.List, k)
	}
	return mapValue

}

func (m MapValue) Len() int {
	return len(m.List)
}

func (m MapValue) Swap(i, j int) {
	m.List[i], m.List[j] = m.List[j], m.List[i]
}

func (m MapValue) Less(i, j int) bool {

	// if the kind of value are not the same, do not need to swap, so return true.

	if !m.Succ {

		return true

	} else if m.Kind == reflect.Float32 {

		return m.Dict[m.List[i]].(float32) < m.Dict[m.List[j]].(float32)

	} else if m.Kind == reflect.Float64 {

		return m.Dict[m.List[i]].(float64) < m.Dict[m.List[j]].(float64)

	} else if m.Kind == reflect.Int {

		return m.Dict[m.List[i]].(int) < m.Dict[m.List[j]].(int)

	} else if m.Kind == reflect.Int8 {

		return m.Dict[m.List[i]].(int8) < m.Dict[m.List[j]].(int8)

	} else if m.Kind == reflect.Int16 {

		return m.Dict[m.List[i]].(int16) < m.Dict[m.List[j]].(int16)

	} else if m.Kind == reflect.Int32 {

		return m.Dict[m.List[i]].(int32) < m.Dict[m.List[j]].(int32)

	} else if m.Kind == reflect.Int64 {

		return m.Dict[m.List[i]].(int64) < m.Dict[m.List[j]].(int64)

	} else {

		return true
	}
}

/*
	This part is used to sort specified map by value of map len.
*/

type MapLen struct {
	Dict map[string]interface{}
	List []string
	Kind reflect.Kind
	Succ bool
}

func ByMapValueLen(m map[string]interface{}) MapLen {

	mapData := MapLen{
		Dict: m,
		Succ: true,
	}

	for k, v := range m {
		if mapData.Kind == reflect.Invalid {
			mapData.Kind = reflect.TypeOf(v).Kind()
		} else if mapData.Kind != reflect.TypeOf(v).Kind() {
			mapData.Succ = false
			break
		}
		mapData.List = append(mapData.List, k)
	}

	if mapData.Kind != reflect.Map && mapData.Kind != reflect.String && mapData.Kind != reflect.Slice {
		mapData.Succ = false
	}

	return mapData
}

func (m MapLen) Len() int {
	return len(m.List)
}

func (m MapLen) Swap(i, j int) {
	m.List[i], m.List[j] = m.List[j], m.List[i]
}

func (m MapLen) Less(i, j int) bool {
	if !m.Succ {
		return true
	} else if m.Kind == reflect.Map || m.Kind == reflect.Slice {
		return reflect.ValueOf(m.Dict[m.List[i]]).Len() < reflect.ValueOf(m.Dict[m.List[j]]).Len()
	} else if m.Kind == reflect.String {
		return len(m.Dict[m.List[i]].(string)) < len(m.Dict[m.List[j]].(string))
	} else {
		return true
	}
}
