package diff

import (
	"context"
	"reflect"
	"testing"
)

func Test_newOutput(t *testing.T) {
	type args struct {
		ctx context.Context
		c   httpClient
		i   input
	}
	tests := []struct {
		name    string
		args    args
		want    output
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newOutput(tt.args.ctx, tt.args.c, tt.args.i, bodyOutput)
			if (err != nil) != tt.wantErr {
				t.Errorf("newOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
