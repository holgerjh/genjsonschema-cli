package merge

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

// MergeAllYAML merges one or more YAML documents into one interface
// The actual type of the returned data depends on the documents
func MergeAllYAML(b ...[]byte) (interface{}, error) {
	raw := make([]interface{}, len(b))
	for i, v := range b {
		if err := yaml.Unmarshal(v, &raw[i]); err != nil {
			return nil, err
		}
	}
	return mergeAll(raw...)
}

func mergeAll(raw ...interface{}) (interface{}, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("expected at least one file")
	}
	result := raw[0]                //default case
	for i := 1; i < len(raw); i++ { // start from second file
		var err error
		result, err = merge(result, raw[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

type jsonType string

const (
	typeObject  jsonType = "object"
	typeArray   jsonType = "array"
	typeString  jsonType = "string"
	typeNumber  jsonType = "number"
	typeInteger jsonType = "integer"
	typeBoolean jsonType = "boolean"
	typeNull    jsonType = "null"
)

func getJSONType(data interface{}) (jsonType, error) {
	switch data.(type) {
	case map[interface{}]interface{}:
		return typeObject, nil
	case map[string]interface{}:
		return typeObject, nil
	case []interface{}:
		return typeArray, nil
	case string:
		return typeString, nil
	case int, int8, int16, int32, int64:
		return typeInteger, nil
	case float32, float64:
		return typeNumber, nil
	case bool:
		return typeBoolean, nil
	case nil:
		return typeNull, nil
	default:
		return typeNull, fmt.Errorf("unexpected type %v of data", reflect.TypeOf(data))
	}
}

// returns true if t2 can be merged into t1
func canMerge(t1, t2 jsonType) bool {
	if t1 == t2 {
		return true
	}
	return (t1 == typeNumber && t2 == typeInteger) || (t1 == typeInteger && t2 == typeNumber)
}

func isScalar(t jsonType) bool {
	return t != typeArray && t != typeObject
}

// mergeScalars merges two scalar values. It is assumed that they are mergeable (-> canMerge)
// this implies that the JSON types are equal or mixed numbers and integers
func mergeScalars(a, b interface{}) (interface{}, error) {
	typeA, err := getJSONType(a)
	if err != nil {
		return nil, err
	}

	// special case: mixed numbers and integers
	// numbers take precendence over integers or otherwise the schema would not accept both types
	if typeA == typeNumber {
		return a, nil
	}

	return b, nil
}

func merge(a, b interface{}) (interface{}, error) {
	typeA, err := getJSONType(a)
	if err != nil {
		return nil, err
	}
	typeB, err := getJSONType(b)
	if err != nil {
		return nil, err
	}

	if !canMerge(typeA, typeB) {
		return nil, fmt.Errorf("rejecting to merge types %s and %s (schema would not accept the given input files)", typeA, typeB)
	}
	// case typeA == typeB

	if isScalar(typeA) {
		return mergeScalars(a, b)
	}
	// -> both are compound values

	if typeA == typeArray {
		// -> both are lists
		return mergeAsLists(a, b)
	} else {
		// -> both are maps
		return mergeAsMaps(a, b)
	}
}

func mergeAsLists(a, b interface{}) ([]interface{}, error) {
	l1, ok1 := a.([]interface{})
	l2, ok2 := b.([]interface{})
	if !(ok1 && ok2) {
		return nil, fmt.Errorf("assumption failed: values are no lists")
	}
	res := make([]interface{}, 0)
	res = append(res, l2...)
	for _, v := range l1 {
		contains := false
		for _, w := range res {
			if reflect.DeepEqual(v, w) {
				contains = true
				break
			}
		}
		if !contains {
			res = append(res, v)
		}
	}
	return res, nil
}

func convertKeysToString(a map[interface{}]interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for k, v := range a {
		stringKey, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("encountered mapping key that is no string")
		}
		res[stringKey] = v
	}
	return res, nil
}

func mergeAsMaps(a, b interface{}) (map[string]interface{}, error) {
	var err error
	var m1, m2 map[string]interface{}
	switch v := a.(type) {
	case map[interface{}]interface{}:
		m1, err = convertKeysToString(v)
		if err != nil {
			return nil, err
		}
	case map[string]interface{}:
		m1 = v
	default:
		return nil, fmt.Errorf("unexpected state")
	}
	switch v := b.(type) {
	case map[interface{}]interface{}:
		m2, err = convertKeysToString(v)
		if err != nil {
			return nil, err
		}
	case map[string]interface{}:
		m2 = v
	default:
		return nil, fmt.Errorf("unexpected state")
	}

	res := make(map[string]interface{})
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		if _, ok := res[k]; ok {
			var err error
			res[k], err = merge(res[k], v) //deepmerge values of keys k
			if err != nil {
				return nil, err
			}
		} else {
			res[k] = v
		}
	}
	return res, nil
}
