package web

import (
	"github.com/tidwall/gjson"
	"strings"
	"testing"
)

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

func TestNormalize(t *testing.T) {
	n1, err := Normalize(json1)
	if err != nil {
		t.Error(err, ": Cannot normalize json1.")
	}
	n2, err := Normalize(json2)
	if err != nil {
		t.Error(err, ": Cannot normalize json2.")
	}
	if !gjson.Valid(n1) {
		t.Error("Normalize function work incorrectly.")
	}
	if strings.Compare(n1, n2) != 0 {
		t.Error("Results are not the same.")
	}
}
