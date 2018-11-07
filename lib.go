package sdm

import "reflect"

// AsArgs converts any array/slice/map to []interface{} panics if not these type
//
// Although passing channel to it does not panic, DO NOT DO THAT!
func AsArgs(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	sz := v.Len()
	ret := make([]interface{}, sz)
	for x := 0; x < sz; x++ {
		ret[x] = v.Index(x).Interface()
	}

	return ret
}

// KeyAsArgs converts any map to []interface{} panics if type mismatch
func KeyAsArgs(arr interface{}) []interface{} {
	keys := reflect.ValueOf(arr).MapKeys()
	ret := make([]interface{}, len(keys))
	for x, k := range keys {
		ret[x] = k.Interface()
	}

	return ret
}
