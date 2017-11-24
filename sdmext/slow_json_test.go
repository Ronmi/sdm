package sdmext

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/Ronmi/sdm"
	"github.com/Ronmi/sdm/driver"
)

func TestSJSONEncode(t *testing.T) {
	ext := &SJSONFactory{}
	now := time.Now()

	cases := []struct {
		name   string
		data   interface{}
		expect string
	}{
		{name: "bool-T", data: true, expect: "true"},
		{name: "bool-F", data: false, expect: "false"},
		{name: "int", data: -1, expect: "-1"},
		{name: "uint", data: 1, expect: "1"},
		{name: "float", data: 1.5, expect: "1.5"},
		{name: "string", data: "a", expect: `"a"`},
		{name: "string-special", data: `a"`, expect: `"a\""`},
		{name: "[]byte", data: []byte(`a"`), expect: `"a\""`},
		{name: "[]rune", data: []rune(`a"`), expect: `"a\""`},
		{name: "time", data: now, expect: strconv.FormatInt(
			now.UnixNano()/1000000, 10)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := reflect.ValueOf(c.data)
			buf := &bytes.Buffer{}
			ext.encode(buf, v)
			if actual := buf.String(); actual != c.expect {
				t.Fatalf("expected [%s], got [%s]", c.expect, actual)
			}
		})
	}
}

type sjsonTestType struct {
	I  int       `sdm:"i" json:"i"`
	U  uint      `sdm:"u" json:"u"`
	F  float64   `sdm:"f" json:"f"`
	B  bool      `sdm:"b" json:"b"`
	S  string    `sdm:"s" json:"s"`
	S2 []byte    `sdm:"s2" json:"s2"`
	S3 []rune    `sdm:"s3" json:"s3"`
	T  time.Time `sdm:"t" json:"t"`
}

func TestSJSONMarshal(t *testing.T) {
	testdata := sjsonTestType{
		-1,
		2,
		3.4,
		true,
		"5",
		[]byte("6"),
		[]rune("7"),
		time.Now(),
	}

	driver.AnsiStub()
	m := sdm.New(nil, "ansistub")
	m.Reg(testdata)

	ext := &SJSONFactory{}
	m.Ext(testdata, ext)

	data, _ := json.Marshal(testdata)
	var expect map[string]interface{}
	var actual map[string]interface{}
	json.Unmarshal(data, &expect)

	data, err := ext.m(testdata)
	if err != nil {
		t.Fatalf("cannot marshal testdata: %s", err)
	}
	if err = json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("unable unmarshal data: %s\n dump data: %s", err, string(data))
	}
}

func BenchmarkSJSONMarshal(b *testing.B) {
	testdata := sjsonTestType{
		-1,
		2,
		3.4,
		true,
		"5",
		[]byte("6"),
		[]rune("7"),
		time.Now(),
	}

	driver.AnsiStub()
	m := sdm.New(nil, "ansistub")
	m.Reg(testdata)

	ext := &SJSONFactory{}
	m.Ext(testdata, ext)

	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		ext.m(testdata)
	}
}

func BenchmarkSJSONMarshalBaseline(b *testing.B) {
	testdata := sjsonTestType{
		-1,
		2,
		3.4,
		true,
		"5",
		[]byte("6"),
		[]rune("7"),
		time.Now(),
	}

	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		json.Marshal(testdata)
	}
}
