package diff

import (
	"strconv"
	"strings"
	"unicode"
)

func parseRows(rows string) []int {
	if rows == "" {
		return []int{}
	}

	parts := strings.Split(rows, ",")
	ret := make([]int, 0, len(parts))
	for _, v := range parts {
		r, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		ret = append(ret, r)
	}

	return ret
}

type sortDelta [][]string

func (s sortDelta) Len() int          { return len(s) }
func (s sortDelta) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortDelta) Less(i int, j int) bool {
	return SortStr(s[i][0], s[j][0])
}

func SortStr(i, j string) bool {
	iRunes := []rune(i)
	jRunes := []rune(j)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}
	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}

func cleanDiff(diff string) string {
	switch diff {
	case "first argument is invalid json":
		return "before API returned invalid json"
	case "second argument is invalid json":
		return "after API returned invalid json"
	default:
		return diff
	}
}
