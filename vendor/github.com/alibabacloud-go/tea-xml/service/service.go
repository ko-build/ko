package service

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"

	"github.com/alibabacloud-go/tea/tea"
	v2 "github.com/clbanning/mxj/v2"
)

func ToXML(obj map[string]interface{}) *string {
	return tea.String(mapToXML(obj))
}

func ParseXml(val *string, result interface{}) map[string]interface{} {
	resp := make(map[string]interface{})

	start := getStartElement([]byte(tea.StringValue(val)))
	if result == nil {
		vm, err := v2.NewMapXml([]byte(tea.StringValue(val)))
		if err != nil {
			return nil
		}
		return vm
	}
	out, err := xmlUnmarshal([]byte(tea.StringValue(val)), result)
	if err != nil {
		return resp
	}
	resp[start] = out
	return resp
}

func mapToXML(val map[string]interface{}) string {
	res := ""
	for key, value := range val {
		switch value.(type) {
		case []interface{}:
			for _, v := range value.([]interface{}) {
				switch v.(type) {
				case map[string]interface{}:
					res += `<` + key + `>`
					res += mapToXML(v.(map[string]interface{}))
					res += `</` + key + `>`
				default:
					if fmt.Sprintf("%v", v) != `<nil>` {
						res += `<` + key + `>`
						res += fmt.Sprintf("%v", v)
						res += `</` + key + `>`
					}
				}
			}
		case map[string]interface{}:
			res += `<` + key + `>`
			res += mapToXML(value.(map[string]interface{}))
			res += `</` + key + `>`
		default:
			if fmt.Sprintf("%v", value) != `<nil>` {
				res += `<` + key + `>`
				res += fmt.Sprintf("%v", value)
				res += `</` + key + `>`
			}
		}
	}
	return res
}

func getStartElement(body []byte) string {
	d := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := d.Token()
		if err != nil {
			return ""
		}
		if t, ok := tok.(xml.StartElement); ok {
			return t.Name.Local
		}
	}
}

func xmlUnmarshal(body []byte, result interface{}) (interface{}, error) {
	start := getStartElement(body)
	dataValue := reflect.ValueOf(result).Elem()
	dataType := dataValue.Type()
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		name, containsNameTag := field.Tag.Lookup("xml")
		name = strings.Replace(name, ",omitempty", "", -1)
		if containsNameTag {
			if name == start {
				realType := dataValue.Field(i).Type()
				realValue := reflect.New(realType).Interface()
				err := xml.Unmarshal(body, realValue)
				if err != nil {
					return nil, err
				}
				return realValue, nil
			}
		}
	}
	return nil, nil
}
