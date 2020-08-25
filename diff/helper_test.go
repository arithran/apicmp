package diff

import (
	"reflect"
	"testing"
)

func Test_Atois(t *testing.T) {
	type args struct {
		rows string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "test 1",
			args: args{
				rows: "1,5,7",
			},
			want: []int{1, 5, 7},
		},
		{
			name: "test 2",
			args: args{
				rows: "",
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Atois(tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Atois() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtoim(t *testing.T) {
	type args struct {
		rows string
	}
	tests := []struct {
		name string
		args args
		want map[int]struct{}
	}{
		{
			name: "test 1",
			args: args{
				rows: "1,5,7",
			},
			want: map[int]struct{}{
				1: {},
				5: {},
				7: {},
			},
		},
		{
			name: "test 2",
			args: args{
				rows: "",
			},
			want: map[int]struct{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Atoim(tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Atoim() = %v, want %v", got, tt.want)
			}
		})
	}
}
