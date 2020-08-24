package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	const json1 = `{
	  "name":         {"first": "Tom", "last": "Anderson"},
	  "age":37,
	  "children": ["Sara","Alex","Jack"],
	  "fav.movie": "Deer Hunter",
	  "friends": [
		{"age": true, "first": "\"Dale", "last": null, "nets": ["ig", "fb", "tw"]},
		{"first": "Roger", "last": "Craig", "age": 68.11, "nets": ["fb", "tw"]},
		{"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	  ]
	}`

	const json2 = `{
	  "fav.movie": "Deer Hunter",
	  "friends": [
		{"first": "\"Dale", "last": null, "age": true, "nets": ["ig", "fb", "tw"]},
		{"first": "Roger", "last": "Craig", "age": 68.11, "nets": ["fb", "tw"]},
		{"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	  ],
	  "age":37,
	  "name": {"first": "Tom", "last": "Anderson"},
	  "children": ["Sara","Alex","Jack"]
	}`

	n1 := Normalize(json1)
	n2 := Normalize(json2)
	fmt.Println(n1)
	if !json.Valid([]byte(n1)) {
		t.Error("Normalize function work incorrectly.")
	}
	if strings.Compare(n1, n2) != 0 {
		t.Error("Results are not the same.")
	}
}
