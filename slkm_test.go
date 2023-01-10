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
		{"Hello @k1low", regexp.MustCompile(`^Hello <@U[0-9A-Z]+>$`)},
	}
	ctx := context.Background()
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := c.replaceMentions(ctx, tt.in)
			if !tt.wantRe.MatchString(got) {
				t.Errorf("got %v\nwant %v", got, tt.wantRe.String())
			}
		})
	}
}

func TestGetChannelIDByName(t *testing.T) {
	if os.Getenv("SLACK_API_TOKEN") == "" {
		t.Skip("not set SLACK_API_TOKEN")
	}
	tests := []struct {
		channel string
		wantRe  *regexp.Regexp
	}{
		{"general", regexp.MustCompile(`^C[0-9A-Z]+$`)},
	}
	ctx := context.Background()
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.channel, func(t *testing.T) {
			got, err := c.getChannelIDByName(ctx, tt.channel)
			if err != nil {
				t.Error(err)
			}
			if !tt.wantRe.MatchString(got) {
				t.Errorf("got %v\nwant %v", got, tt.wantRe.String())
			}
		})
	}
}
