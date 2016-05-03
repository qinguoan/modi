package reaper

import (
	"fmt"
	"testing"
	//"time"
)

/*
func TestGetDomainPath(t *testing.T) {
	testChan := getDomainPath()
	select {
	case <-time.After(time.Second * 5):
		t.Fatal("couldn't receive data from DomainBindUrl in 10 seconds")
	case <-testChan:
		t.Log("receive data from DomainBindUrl OK")
	}
}
*/
/*
func TestStartService(t *testing.T) {
	go StartService()
	time.Sleep(time.Second * 15)
	// fmt.Println(curBindData)
	t.Log(curBindData)
}
*/

func TestAggregationPath(t *testing.T) {
	a := map[string]int{
		"/v3/specialCategory":              1,
		"/xxxsdsdsfreetert123_12312312444": 738398,
	}

	c := AggregationPath(a)

	fmt.Println(c)
}

/*
func TestDeepEqual(t *testing.T) {
	a := map[string][]string{
		"www.baidu.com":     []string{"a", "b", "c"},
		"www.wandoujia.com": []string{"/api/v1", "/api/v2"},
		"www.snapea.com":    []string{"/v2", "/v1", "/v3"},
	}

	b := map[string][]string{
		"www.wandoujia.com": []string{"/api/v1", "/api/v2"},
		"www.snapea.com":    []string{"/v2", "/v1", "/v3"},
		"www.baidu.com":     []string{"c", "b", "a"},
	}

	eq :=

	if !eq {
		t.Fatal("wrong result")
	}
}
*/
