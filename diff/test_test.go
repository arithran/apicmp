package diff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "Test GET",
			args: args{
				c: Config{
					BeforeBasePath:  "http://before.api.com",
					AfterBasePath:   "http://after.api.com",
					FixtureFilePath: "./testdata/get.csv",
				},
			},
			want: []test{
				{
					Row: 1,
					Before: input{
						Method: "GET",
						Path:   "http://before.api.com/users/1",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.1",
							"Content-Type":    "application/json",
						},
					},
					After: input{
						Method: "GET",
						Path:   "http://after.api.com/users/1",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.1",
							"Content-Type":    "application/json",
						},
					},
				},
				{
					Row: 2,
					Before: input{
						Method: "GET",
						Path:   "http://before.api.com/users/2",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.2",
							"Content-Type":    "application/json",
						},
					},
					After: input{
						Method: "GET",
						Path:   "http://after.api.com/users/2",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.2",
							"Content-Type":    "application/json",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test POST",
			args: args{
				c: Config{
					BeforeBasePath:  "http://before.api.com",
					AfterBasePath:   "http://after.api.com",
					FixtureFilePath: "./testdata/post.csv",
				},
			},
			want: []test{
				{
					Row: 1,
					Before: input{
						Method: "POST",
						Path:   "http://before.api.com/users/create",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.1",
							"Content-Type":    "application/json",
						},
						Body: `{"email": "user1@example.com", "password": "pa$$word"}`,
					},
					After: input{
						Method: "POST",
						Path:   "http://after.api.com/users/create",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.1",
							"Content-Type":    "application/json",
						},
						Body: `{"email": "user1@example.com", "password": "pa$$word"}`,
					},
				},
				{
					Row: 2,
					Before: input{
						Method: "POST",
						Path:   "http://before.api.com/users/create",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.2",
							"Content-Type":    "application/json",
						},
						Body: `{"email": "user2@example.com", "password": "pa$$word"}`,
					},
					After: input{
						Method: "POST",
						Path:   "http://after.api.com/users/create",
						Headers: map[string]string{
							"X-Api-Key":       "abcd",
							"X-Forwarded-For": "192.168.1.2",
							"Content-Type":    "application/json",
						},
						Body: `{"email": "user2@example.com", "password": "pa$$word"}`,
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

			assert.Equal(t, tt.want, got)
		})
	}
}
