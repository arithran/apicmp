package diff

import (
	"errors"
	"plugin"
	"regexp"

	"github.com/arithran/apicmp/module"
)

var (
	arrayRegexp = regexp.MustCompile(`\[([0-9*]+)\]`)
)

type valueType int

const (
	noneValueType valueType = 0
	strValueType  valueType = 1
)

func configureWithModule(modulePath string) (module.RawCompare, error) {
	plug, err := plugin.Open(modulePath)
	if err != nil {
		return nil, err
	}
	symGetConfig, err := plug.Lookup("Cmp")
	if err != nil {
		return nil, err
	}

	cmp, ok := symGetConfig.(*module.RawCompare)
	if !ok {
		return nil, errors.New("Cmp has no proper type")
	}
	return *cmp, nil
}
