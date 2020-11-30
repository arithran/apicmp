package diff

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

// Atois converts Ascii csv to int slice
func Atois(csv string) []int {
	if csv == "" {
		return []int{}
	}

	parts := strings.Split(csv, ",")
	ret := make([]int, 0, len(parts))
	for _, v := range parts {
		r, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("Atois err:%v", err)
			continue
		}
		ret = append(ret, r)
	}
	return ret
}

// Atoim converts Ascii csv to int map
func Atoim(csv string) map[int]struct{} {
	ret := make(map[int]struct{})
	if csv == "" {
		return ret
	}

	parts := strings.Split(csv, ",")
	for _, v := range parts {
		r, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("Atoim err:%v", err)
			continue
		}
		ret[r] = struct{}{}
	}
	return ret
}

// Atoam converts Ascii csv to Ascii map
func Atoam(csv string) map[string]struct{} {
	ret := make(map[string]struct{})
	if csv == "" {
		return ret
	}

	parts := strings.Split(csv, ",")
	for _, v := range parts {
		ret[v] = struct{}{}
	}
	return ret
}

// Istoa converts a slice of int to Ascii
func Istoa(slice []int, sep string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(slice)), sep), "[]")
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

func setLoglevel(level string) error {
	l, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	log.SetLevel(l)
	return nil
}

func buildURL(base, path string, qs []string, ignoreQS *regexp.Regexp) string {
	out := base + path

	if len(qs) > 0 || ignoreQS != nil {
		u, _ := url.Parse(out)
		q, _ := url.ParseQuery(u.RawQuery)

		// delete query strings that match the regex
		if ignoreQS != nil {
			for k := range q {
				if ignoreQS.MatchString(k) {
					q.Del(k)
				}
			}
		}

		for _, h := range qs {
			parts := strings.Split(h, ":")
			if len(parts) != headerParts {
				log.Errorf("skipping invalid param --querystring %s", h)
				continue
			}

			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			q.Add(k, v)
		}

		u.RawQuery = q.Encode()
		out = u.String()
	}

	return out
}
