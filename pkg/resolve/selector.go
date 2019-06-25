package resolve

import (
	"bytes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	yaml2json "k8s.io/apimachinery/pkg/util/yaml"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
)

const yamlSeparator = "\n---"

var yamlSeparatorRegex = regexp.MustCompile(yamlSeparator)

// FilterBySelector filters out any resources
// from the raw manifest bytes whose labels
// don't match the provided selector
func FilterBySelector(input []byte, selectorString string) ([]byte, error) {
	selector, err := labels.Parse(selectorString)
	if err != nil {
		return nil, err
	}

	objects, err := parseUnstructured(input)
	if err != nil {
		return nil, err
	}

	var rawObjYamls [][]byte
	encodeObj := func(obj runtime.Object) error {
		rawJson, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
		if err != nil {
			return err
		}
		rawYaml, err := yaml.JSONToYAML(rawJson)
		if err != nil {
			return err
		}
		rawObjYamls = append(rawObjYamls, rawYaml)
		return nil
	}
	for _, object := range objects {
		switch unstructuredObj := object.(type) {
		case *unstructured.Unstructured:
			if selector.Matches(labels.Set(unstructuredObj.GetLabels())) {
				if err := encodeObj(unstructuredObj); err != nil {
					return nil, err
				}
			}
		case *unstructured.UnstructuredList:
			var filteredItems []unstructured.Unstructured
			for _, obj := range unstructuredObj.Items {
				if selector.Matches(labels.Set(obj.GetLabels())) {
					filteredItems = append(filteredItems, obj)
				}
			}
			if len(filteredItems) > 0 {
				unstructuredObj.Items = filteredItems
				if err := encodeObj(unstructuredObj); err != nil {
					return nil, err
				}
			}
		}
	}

	return bytes.Join(rawObjYamls, []byte("---\n")), nil
}

func parseUnstructured(rawYaml []byte) ([]runtime.Object, error) {
	objectYamls := yamlSeparatorRegex.Split(string(rawYaml), -1)
	var resources []runtime.Object

	for _, objectYaml := range objectYamls {
		if isEmptyYamlSnippet(objectYaml) {
			continue
		}
		jsn, err := yaml2json.ToJSON([]byte(objectYaml))
		if err != nil {
			return nil, err
		}
		runtimeObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsn)
		if err != nil {
			return nil, err
		}

		resources = append(resources, runtimeObj)
	}

	return resources, nil
}

var commentRegex = regexp.MustCompile("#.*")

func isEmptyYamlSnippet(manifest string) bool {
	removeComments := commentRegex.ReplaceAllString(manifest, "")
	removeNewlines := strings.Replace(removeComments, "\n", "", -1)
	removeDashes := strings.Replace(removeNewlines, "---", "", -1)
	removeSpaces := strings.Replace(removeDashes, " ", "", -1)
	return removeSpaces == ""
}
