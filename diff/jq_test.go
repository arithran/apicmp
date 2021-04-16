package diff

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jqMatchesToBody(t *testing.T) {
	type args struct {
		result []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]json.RawMessage
		wantErr bool
	}{
		{
			name: "matches list is nil",
			args: args{
				result: nil,
			},
			want: map[string]json.RawMessage{
				"jq": []byte("null"),
			},
		},
		{
			name: "matches list is empty",
			args: args{
				result: []interface{}{},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("null"),
			},
		},
		{
			name: "matches list contains a single object",
			args: args{
				result: []interface{}{
					map[string]interface{}{
						"null":   nil,
						"bool":   true,
						"number": 42.42,
						"string": "artes serviunt vitae, sapientia imperat",
						"list":   []interface{}{true, 42.42, "some string"},
						"object": map[string]interface{}{"anything": 124},
					},
				},
			},
			want: map[string]json.RawMessage{
				"jq:null":   []byte("null"),
				"jq:bool":   []byte("true"),
				"jq:number": []byte("42.42"),
				"jq:string": []byte("\"artes serviunt vitae, sapientia imperat\""),
				"jq:list":   []byte("[true,42.42,\"some string\"]"),
				"jq:object": []byte("{\"anything\":124}"),
			},
		},
		{
			name: "matches list contains a single string",
			args: args{
				result: []interface{}{
					"artes serviunt vitae, sapientia imperat",
				},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("\"artes serviunt vitae, sapientia imperat\""),
			},
		},
		{
			name: "matches list contains a single number",
			args: args{
				result: []interface{}{
					42.42,
				},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("42.42"),
			},
		},
		{
			name: "matches list contains a single bool",
			args: args{
				result: []interface{}{
					true,
				},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("true"),
			},
		},
		{
			name: "matches list contains a single list",
			args: args{
				result: []interface{}{
					[]interface{}{true, 42.42, "some string"},
				},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("[true,42.42,\"some string\"]"),
			},
		},
		{
			name: "matches list contains a list of a values",
			args: args{
				result: []interface{}{
					nil,
					true,
					42.42,
					"artes serviunt vitae, sapientia imperat",
					[]interface{}{true, 42.42, "some string"},
					map[string]interface{}{"anything": 124},
				},
			},
			want: map[string]json.RawMessage{
				"jq": []byte("[null,true,42.42,\"artes serviunt vitae, sapientia imperat\",[true,42.42,\"some string\"],{\"anything\":124}]"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jqMatchesToBody(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("jqMatchesToBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}
