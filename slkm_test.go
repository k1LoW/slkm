package slkm

import (
	"context"
	"os"
	"regexp"
	"testing"
)

func TestReplaceMentions(t *testing.T) {
	if os.Getenv("SLACK_API_TOKEN") == "" {
		t.Skip("not set SLACK_API_TOKEN")
	}
	tests := []struct {
		in     string
		wantRe *regexp.Regexp
	}{
		{"Hello @here", regexp.MustCompile("Hello <!here>")},
		{"Hello @k1low", regexp.MustCompile("Hello <@[0-9A-Z]+>")},
	}
	for _, tt := range tests {
		ctx := context.Background()
		c, err := New()
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.in, func(t *testing.T) {
			got := c.replaceMentions(ctx, tt.in)
			if !tt.wantRe.MatchString(got) {
				t.Errorf("got %v\nwant %v", got, tt.wantRe.String())
			}
		})
	}
}
