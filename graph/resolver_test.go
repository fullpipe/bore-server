//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"context"
	"testing"
)

func Test_getFileDuration(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    float64
		wantErr bool
	}{
		{
			"it fails on invalid path",
			"foo",
			0,
			true,
		},
		{
			"it fails on empty path",
			"foo",
			0,
			true,
		},
		{
			"it returns valid duration",
			"../var/test.mp3",
			524.669388,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFileDuration(context.Background(), tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getFileDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
