package diff

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseRow(t *testing.T) {
	type args struct {
		row    int
		inp    input
		suffix string
	}
	tests := []struct {
		name string
		args args
		want Item
	}{
		{
			name: "GET request",
			args: args{
				row: 1,
				inp: input{
					Method: "GET",
					Path:   "http://localhost:8080/model/id?k=v",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: "",
				},
				suffix: "Before",
			},
			want: Item{
				Name: "Row 1 - Before",
				ProtocolProfileBehavior: ProtocolProfileBehavior{
					DisableBodyPruning: true,
				},
				Request: Request{
					Method: "GET",
					Header: []Header{
						{
							Key:   "Content-Type",
							Value: "application/json",
						},
					},
					Body: Body{
						Mode: "raw",
						Raw:  "",
					},
					URL: URL{
						Raw:      "",
						Protocol: "http",
						Host:     []string{"localhost:8080"},
						Path:     []string{"", "model", "id"},
						Query:    []Query{{Key: "k", Value: "v"}},
					},
				},
			},
		},
		{
			name: "POST request",
			args: args{
				row: 1,
				inp: input{
					Method: "POST",
					Path:   "http://localhost:8080/model/id?k=v",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: "{}",
				},
				suffix: "After",
			},
			want: Item{
				Name: "Row 1 - After",
				ProtocolProfileBehavior: ProtocolProfileBehavior{
					DisableBodyPruning: true,
				},
				Request: Request{
					Method: "POST",
					Header: []Header{
						{
							Key:   "Content-Type",
							Value: "application/json",
						},
					},
					Body: Body{
						Mode: "raw",
						Raw:  "{}",
					},
					URL: URL{
						Raw:      "",
						Protocol: "http",
						Host:     []string{"localhost:8080"},
						Path:     []string{"", "model", "id"},
						Query:    []Query{{Key: "k", Value: "v"}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRow(tt.args.row, tt.args.inp, tt.args.suffix); !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
