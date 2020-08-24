package diff

import (
	"reflect"
	"testing"
)

func Test_parseRows(t *testing.T) {
	type args struct {
		rows string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "test1",
			args: args{
				rows: "1,5,7",
			},
			want: []int{1, 5, 7},
		},
		{
			name: "test2",
			args: args{
				rows: "",
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRows(tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRows() = %v, want %v", got, tt.want)
			}
		})
	}
}
