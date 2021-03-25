package main

import (
	"fmt"
	"reflect"
	"strings"
)

func main() {
	queryString := `{
  topicsById(ids: ["x2f8bb11595b61c86"]) {
    id
    slug
    title
    contentKind
    parentTopicId
    unitIds # deprecated
    filteredContent(kinds: ["Exercise"]) {
      id
      mappedStandards {
        standardId
        setId
      }
      contentKind
      parentTopic{
        id
        mappedStandardIds
      }
    }
  }
}`
	var query struct {
		District struct {
			ID      string
			Schools []struct {
				ID   string
				Name string
			}
		} `graphql:"query"`
	}
	PrintFields(query)
	taggedQuery := tagQueryStruct(queryString, query)

	PrintFields(taggedQuery)
}

func tagQueryStruct(queryString string, query interface{}) interface{} {
	queryString = standardizeSpaces(queryString)
	// get existing field tag
	pt := reflect.TypeOf(query)

	var st []reflect.StructField
	for i := 0; i < pt.NumField(); i++ {
		st = append(st, pt.Field(i))
	}
	v := reflect.ValueOf(query)
	if len(st) == 0 {
		return query
	}
	newTag := &Tag{
		Key:     "graphql",
		Name:    queryString,
		Options: []string{},
	}
	// replace with our modified tag
	st[0].Tag = reflect.StructTag(newTag.String())

	p2 := reflect.StructOf(st)

	v2 := v.Convert(p2)
	return v2.Interface()
}

func PrintFields(b interface{}) {
	val := reflect.ValueOf(b)
	for i := 0; i < val.Type().NumField(); i++ {
		fmt.Println(val.Type().Field(i).Tag.Get("graphql"))
	}
}

// Value returns the raw value of the tag, i.e. if the tag is
// `json:"foo,omitempty", the Value is "foo,omitempty"
func (t *Tag) Value() string {
	options := strings.Join(t.Options, ",")
	if options != "" {
		return fmt.Sprintf(`%s,%s`, t.Name, options)
	}
	return t.Name
}

func (t *Tag) String() string {
	return fmt.Sprintf(`%s:%q`, t.Key, t.Value())
}

// Tag defines a single struct's string literal tag
type Tag struct {
	// Key is the tag key, such as json, xml, etc..
	// i.e: `json:"foo,omitempty". Here key is: "json"
	Key string

	// Name is a part of the value
	// i.e: `json:"foo,omitempty". Here name is: "foo"
	Name string

	// Options is a part of the value. It contains a slice of tag options i.e:
	// `json:"foo,omitempty". Here options is: ["omitempty"]
	Options []string
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
