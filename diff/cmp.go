package diff

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/arithran/apicmp/module"
	"github.com/arithran/jsondiff"
)

type outputComparer interface {
	cmp(before, after output) (diff, bool)
	outputType() outputType
}

type jsondiffCmp struct {
	ignore    map[string]struct{}
	opts      jsondiff.Options
	wantMatch jsondiff.Difference
}

func newJSONDiffCmp(ignore map[string]struct{}, opts jsondiff.Options, wantMatch jsondiff.Difference) jsondiffCmp {
	return jsondiffCmp{
		ignore:    ignore,
		opts:      opts,
		wantMatch: wantMatch,
	}
}

func (jsDiff jsondiffCmp) cmp(before, after output) (diff, bool) {
	for k, v := range before.Body {
		if _, ok := jsDiff.ignore[k]; ok {
			continue
		}

		match, delta := jsondiff.Compare(after.Body[k], v, &jsDiff.opts)
		if match > jsDiff.wantMatch {
			return diff{
				Field: k,
				Delta: cleanDiff(delta),
			}, false
		}
	}
	return diff{}, true
}

func (jsDiff jsondiffCmp) outputType() outputType {
	return bodyOutput
}

type pluginCmp struct {
	compareNative module.RawCompare
}

func newPluginCmp(moduleFilePath string) (pluginCmp, error) {
	plug, err := plugin.Open(moduleFilePath)
	if err != nil {
		return pluginCmp{}, err
	}
	symGetConfig, err := plug.Lookup("Cmp")
	if err != nil {
		return pluginCmp{}, err
	}

	rawCmp, ok := symGetConfig.(*module.RawCompare)
	if !ok || rawCmp == nil {
		return pluginCmp{}, errors.New("Cmp has no proper type")
	}
	return pluginCmp{
		compareNative: *rawCmp,
	}, nil
}

func (plug pluginCmp) cmp(before, after output) (diff, bool) {
	d, ok := plug.compareNative(before.Raw, after.Raw)
	if ok {
		return diff{}, true
	}
	return diff{
		Field: d.Path,
		Delta: fmt.Sprintf("%s => %s", d.Val1, d.Val2),
	}, false
}

func (plug pluginCmp) outputType() outputType {
	return rawOutput
}
