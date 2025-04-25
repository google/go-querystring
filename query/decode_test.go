package query

import (
	"net/url"
	"reflect"
	"testing"
)

func testDecode(t *testing.T, values url.Values, target interface{}, want interface{}) {
	err := Decode(values, target)
	if err != nil {
		t.Errorf("Decode() method error %v", err)
	}
	obj := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if !reflect.DeepEqual(obj, want) {
		t.Errorf("Decode() method result is %#v; want %#v", obj, want)
	}
}

func TestDecode_BasicTypes(t *testing.T) {
	type TestData struct {
		CompanyName string `url:"company_name"`
		Employees   int    `url:"employees"`
		IsFaang     bool   `url:"is_faang"`
	}

	vals := url.Values{
		"company_name": {"Google"},
		"employees":    {"180000"},
		"is_faang":     {"true"},
	}

	var result TestData
	want := TestData{CompanyName: "Google", Employees: 180000, IsFaang: true}

	testDecode(t, vals, &result, want)
}
