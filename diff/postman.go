package diff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type (
	Collection struct {
		Info Info   `json:"info"`
		Item []Item `json:"item"`
	}
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	}
	Item struct {
		Name                    string                  `json:"name"`
		ProtocolProfileBehavior ProtocolProfileBehavior `json:"protocolProfileBehavior"`
		Request                 Request                 `json:"request"`
	}
	ProtocolProfileBehavior struct {
		DisableBodyPruning bool `json:"disableBodyPruning"`
	}
	Request struct {
		Method string   `json:"method"`
		Header []Header `json:"header"`
		Body   Body     `json:"body"`
		URL    URL      `json:"url"`
	}
	Header struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	Body struct {
		Mode string `json:"mode"`
		Raw  string `json:"raw"`
	}
	URL struct {
		Raw      string   `json:"raw"`
		Protocol string   `json:"protocol"`
		Host     []string `json:"host"`
		Path     []string `json:"path"`
		Query    []Query  `json:"query"`
	}
	Query struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

type Postman interface {
	GenerateCollection(filePath string, ts []test)
}

type PostmanV2 struct{}

func (p PostmanV2) GenerateCollection(path string, ts []test) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	item := make([]Item, 0)
	for _, t := range ts {
		item = append(item, parseRow(t.Row, t.Before, "Before"))
		item = append(item, parseRow(t.Row, t.After, "After"))
	}

	list := strings.Split(f.Name(), string(os.PathSeparator))
	col := Collection{
		Info: Info{
			Name:   fmt.Sprintf("%s__%s", list[len(list)-1], time.Now().Format("2006-01-02T15:04:05")),
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: item,
	}
	buf, _ := json.Marshal(col)
	w := bufio.NewWriter(f)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return w.Flush()
}

func parseRow(row int, inp input, suffix string) Item {
	u, _ := url.Parse(inp.Path)
	qs := make([]Query, 0)
	for k, vs := range u.Query() {
		for _, v := range vs {
			qs = append(qs, Query{
				Key:   k,
				Value: v,
			})
		}
	}

	hs := make([]Header, 0)
	for k, v := range inp.Headers {
		hs = append(hs, Header{Key: k, Value: v})
	}

	return Item{
		Name: fmt.Sprintf("Row %d - %s", row, suffix),
		ProtocolProfileBehavior: ProtocolProfileBehavior{
			DisableBodyPruning: true,
		},
		Request: Request{
			Method: inp.Method,
			Header: hs,
			Body: Body{
				Mode: "raw",
				Raw:  inp.Body,
			},
			URL: URL{
				Raw:      u.RawPath,
				Protocol: u.Scheme,
				Host:     []string{u.Host},
				Path:     strings.Split(u.Path, "/"),
				Query:    qs,
			},
		},
	}
}
