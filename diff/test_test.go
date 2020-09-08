package diff

import (
	"context"
	"testing"

	as "github.com/stretchr/testify/assert"
)

const (
	headerAPIKey  = "X-Api-Key"
	headerUserDma = "X-User-Dma"
)

func Test_generateTests(t *testing.T) {
	type args struct {
		c Config
	}
	tests := []struct {
		name    string
		args    args
		want    []test
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				c: Config{
					BeforeBasePath:  "http://before.api.com",
					AfterBasePath:   "http://after.api.com",
					FixtureFilePath: "./testdata/test1.csv",
				},
			},
			want: []test{
				{
					Row: 1,
					Before: input{
						Method: "GET",
						Path:   "http://before.api.com/video/2387e4d6a7bede9342150d9afbd0d20f",
						Headers: map[string]string{
							headerAPIKey:  "1234abcd",
							headerUserDma: "999",
						},
					},
					After: input{
						Method: "GET",
						Path:   "http://after.api.com/video/2387e4d6a7bede9342150d9afbd0d20f",
						Headers: map[string]string{
							headerAPIKey:  "1234abcd",
							headerUserDma: "999",
						},
					},
				},
				{
					Row: 2,
					Before: input{
						Method: "GET",
						Path:   "http://before.api.com/video/3e3a3ecbf14f85db2c74a3b79452f3f1",
						Headers: map[string]string{
							headerAPIKey:  "1234abcd",
							headerUserDma: "636",
						},
					},
					After: input{
						Method: "GET",
						Path:   "http://after.api.com/video/3e3a3ecbf14f85db2c74a3b79452f3f1",
						Headers: map[string]string{
							headerAPIKey:  "1234abcd",
							headerUserDma: "636",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testChan, err := generateTests(context.Background(), tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := []test{}
			for t := range testChan {
				got = append(got, t)
			}

			as.Equal(t, tt.want, got)
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("generateTests() = %v, want %v", got, tt.want)
			// }
		})
	}
}
