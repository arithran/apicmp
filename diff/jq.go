package diff

import (
	"encoding/json"

	"github.com/itchyny/gojq"
)

// jqMatchesToBody converts a list of jq matches to a diff-able object.
// jq matches might not be a valid json, which would not be diffed correctly.
// The following cases are handled:
// - A nil or empty match gets converted to {"jq": null}.
// - A single match containing a single object gets diffed by property, with a "jq:" prefix.
// - A match or list of matches of any other types (lists, numbers, strings) get diffed as a single "jq" property.
func jqMatchesToBody(result []interface{}) (map[string]json.RawMessage, error) {
	var cmp interface{}
	switch len(result) {
	// null
	case 0:
		return map[string]json.RawMessage{
			"jq": []byte("null"),
		}, nil
	// object or single element
	case 1:
		cmp = result[0]
		// return early if it's an object
		obj, ok := cmp.(map[string]interface{})
		if ok {
			// encode every field sepearately, prefixed "jq:"
			enc := make(map[string]json.RawMessage, len(obj))
			for k, v := range obj {
				val, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				enc["jq:"+k] = val
			}
			return enc, nil
		}
	// list of elements
	default:
		cmp = result
	}

	// just encode it as json and assign as a special field
	enc, err := json.Marshal(cmp)
	if err != nil {
		return nil, err
	}
	return map[string]json.RawMessage{
		"jq": enc,
	}, nil
}

func runJqQuery(jq *gojq.Query, body interface{}) ([]interface{}, error) {
	iter := jq.Run(body)
	res := []interface{}{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func applyJqQueryToBody(jq *gojq.Query, body interface{}) (map[string]json.RawMessage, error) {
	res, err := runJqQuery(jq, body)
	if err != nil {
		return nil, err
	}
	return jqMatchesToBody(res)
}
