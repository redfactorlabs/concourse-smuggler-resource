package smuggler

import (
	"encoding/json"
)

// Check if the string is json itself, in which case is parsed and
// return as interface{}. If not, returns the string itself
func JsonStringToInterface(s string) interface{} {
	var r interface{}
	b := []byte(s)

	err := json.Unmarshal(b, &r)
	if err == nil {
		return r
	}
	return s
}

func JsonStringToInterfaceList(sl []string) []interface{} {
	vl := []interface{}{}
	for _, s := range sl {
		vl = append(vl, JsonStringToInterface(s))
	}
	return vl
}

func InterfaceToJsonString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	default:
	}
	s, err := json.Marshal(v)
	if err != nil {
		panic("Error converting version to json. Shouldn't happen :(")
	}
	return string(s)
}
