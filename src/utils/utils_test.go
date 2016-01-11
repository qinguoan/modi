package utils

import (
	"fmt"
	//"strings"
	"testing"
)

/*
func TestFindConfigPath(t *testing.T) {
	ss := []string{"/"}
	s := "/v3/specialCategroy"

	final, _ := FindConfigPath(s, ss)
	fmt.Println(final)

}
*/

func TestCurrentStamp(t *testing.T) {

	a, _ := CurrentStamp(60)

	fmt.Println(a)
}

func TestCheckLegalChar(t *testing.T) {
	a := map[string]string{"upstrean": "10.0.66.21:8080, "}
	fmt.Println(CheckLegalChar(a))
}

/*
func TestChangeToLegalString(t *testing.T) {
	a := changeToLegalString("/five/\xF61/search")
	fmt.Println(a)
}

func TestCutHead(t *testing.T) {
	a := "http://xxxxx/xx"
	fmt.Println(strings.Index(a, "http://"))
	a = CutHttpHead(a)
	fmt.Println(a)
}
*/
