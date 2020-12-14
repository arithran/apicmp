package diff

import (
	"reflect"
	"regexp"
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

func Test_buildURL(t *testing.T) {
	type args struct {
		base     string
		path     string
		qs       []string
		ignoreQS *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test 1",
			args: args{
				base: "http://localhost",
				path: "/users/1",
				qs: []string{
					"bar: tom",
					"baz: dick",
				},
				ignoreQS: nil,
			},
			want: "http://localhost/users/1?bar=tom&baz=dick",
		},
		{
			name: "test 2",
			args: args{
				base: "http://localhost",
				path: "/users/1?foo=true",
				qs: []string{
					"bar: tom",
					"baz: dick",
				},
				ignoreQS: nil,
			},
			want: "http://localhost/users/1?bar=tom&baz=dick&foo=true",
		},
		{
			name: "test 3",
			args: args{
				base: "http://localhost",
				path: "/users/1?foo=true",
				qs: []string{
					"bar: tom",
				},
				ignoreQS: regexp.MustCompile("foo"),
			},
			want: "http://localhost/users/1?bar=tom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildURL(tt.args.base, tt.args.path, tt.args.qs, tt.args.ignoreQS); got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
