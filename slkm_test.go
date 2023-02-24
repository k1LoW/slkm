package slkm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/slack-go/slack"
)

func TestReplaceBlockMentions(t *testing.T) {
	if os.Getenv("SLACK_API_TOKEN") == "" {
		t.Skip("not set SLACK_API_TOKEN")
	}
	tests := []struct {
		in slack.Block
	}{
		{slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "hello @here", false, false), nil, nil)},
		{slack.NewSectionBlock(nil, []*slack.TextBlockObject{slack.NewTextBlockObject("mrkdwn", "hello @k1low", false, false)}, nil)},
	}
	ctx := context.Background()
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			a, err := json.Marshal(tt.in)
			if err != nil {
				t.Error(err)
			}
			if err := c.replaceBlockMentions(ctx, tt.in); err != nil {
				t.Error(err)
			}
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(a, b, nil); diff == "" {
				t.Error("want diff")
			}
		})
	}

}

func TestReplaceMentionsToMentionLinks(t *testing.T) {
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
			got := c.ReplaceMentionsToMentionLinks(ctx, tt.in)
			if !tt.wantRe.MatchString(got) {
				t.Errorf("got %v\nwant %v", got, tt.wantRe.String())
			}
		})
	}
}

func TestFindChannelIDByName(t *testing.T) {
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
			got, err := c.FindChannelIDByName(ctx, tt.channel)
			if err != nil {
				t.Error(err)
			}
			if !tt.wantRe.MatchString(got) {
				t.Errorf("got %v\nwant %v", got, tt.wantRe.String())
			}
		})
	}
}
