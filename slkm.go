package slkm

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
)

type Client struct {
	client         *slack.Client
	channelCache   map[string]slack.Channel
	userCache      map[string]slack.User
	userGroupCache map[string]slack.UserGroup

	username  string
	iconEmoji string
	iconURL   string

	token      string
	webhookURL string
}

func New() (*Client, error) {
	token := os.Getenv("SLACK_API_TOKEN")
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	c := &Client{
		client:         slack.New(token),
		channelCache:   map[string]slack.Channel{},
		userCache:      map[string]slack.User{},
		userGroupCache: map[string]slack.UserGroup{},
		token:          token,
		webhookURL:     webhookURL,
	}
	return c, nil
}

// PostMessage to Slack channel
func (c *Client) PostMessage(ctx context.Context, channel string, blocks ...slack.Block) error {
	if c.token == "" && c.webhookURL != "" {
		return slack.PostWebhookContext(ctx, c.webhookURL, &slack.WebhookMessage{
			Username:  c.username,
			IconEmoji: c.iconEmoji,
			IconURL:   c.iconURL,
			Channel:   channel,
			Blocks:    &slack.Blocks{BlockSet: blocks},
		})
	}
	channelID, err := c.FindChannelIDByName(ctx, channel)
	if err != nil {
		return err
	}

	for _, b := range blocks {
		if err := c.replaceBlockMentions(ctx, b); err != nil {
			return err
		}
	}

	opts := []slack.MsgOption{
		slack.MsgOptionBlocks(blocks...),
	}
	if c.username != "" {
		opts = append(opts, slack.MsgOptionUsername(c.username))
	}
	if c.iconEmoji != "" {
		opts = append(opts, slack.MsgOptionIconEmoji(c.iconEmoji))
	}
	if c.iconURL != "" {
		opts = append(opts, slack.MsgOptionIconURL(c.iconURL))
	}

	if _, _, err := c.client.PostMessageContext(ctx, channelID, opts...); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetUsername(username string) {
	c.username = username
}

func (c *Client) SetIconEmoji(emoji string) {
	c.iconEmoji = emoji
}

func (c *Client) SetIconURL(u string) {
	c.iconURL = u
}

func (c *Client) SetToken(token string) {
	c.token = token
	c.client = slack.New(token)
}

func (c *Client) SetWebhookURL(url string) {
	c.webhookURL = url
}

func (c *Client) SlackClient() *slack.Client {
	return c.client
}

var mentionRe = regexp.MustCompile(`@[^\s@]+`)

// ReplaceMentionsToMentionLinks replace mentions ( `@k1low` ) to mention links ( `<@UXXXXXXX>` )
func (c *Client) ReplaceMentionsToMentionLinks(ctx context.Context, in string) string {
	mentions := mentionRe.FindAllString(in, -1)
	oldnew := []string{}
	for _, m := range mentions {
		l, err := c.CreateMentionLinkFromName(ctx, m)
		if err != nil {
			continue
		}
		oldnew = append(oldnew, m, l)
	}
	rep := strings.NewReplacer(oldnew...)
	return rep.Replace(in)
}

// FindChannelIDByName find Channel ID ( `CXXXXXXX` ) by channel name ( `#general` )
func (c *Client) FindChannelIDByName(ctx context.Context, channel string) (string, error) {
	channel = strings.TrimPrefix(channel, "#")
	if cc, ok := c.channelCache[channel]; ok {
		return cc.ID, nil
	}
	var (
		nc  string
		err error
		cID string
	)
L:
	for {
		var ch []slack.Channel
		p := &slack.GetConversationsParameters{
			Limit:  1000,
			Cursor: nc,
		}
		ch, nc, err = c.client.GetConversationsContext(ctx, p)
		if err != nil {
			return "", err
		}
		for _, cc := range ch {
			c.channelCache[channel] = cc
			if cc.Name == channel {
				cID = cc.ID
				break L
			}
		}
		if nc == "" {
			break
		}
	}
	if cID == "" {
		return "", fmt.Errorf("not found channel: %s", channel)
	}

	return cID, nil
}

// CreateMentionLinkFromName create mention link ( `<@UXXXXXXX>` ) from mention ( `@k1low` )
func (c *Client) CreateMentionLinkFromName(ctx context.Context, name string) (string, error) {
	name = strings.TrimPrefix(name, "@")
	switch name {
	case "channel", "here", "everyone":
		return fmt.Sprintf("<!%s>", name), nil
	}
	if uc, ok := c.userCache[name]; ok {
		// https://api.slack.com/reference/surfaces/formatting#mentioning-users
		return fmt.Sprintf("<@%s>", uc.ID), nil
	}
	if gc, ok := c.userGroupCache[name]; ok {
		// https://api.slack.com/reference/surfaces/formatting#mentioning-groups
		return fmt.Sprintf("<!subteam^%s>", gc.ID), nil
	}

	users, err := c.client.GetUsersContext(ctx)
	if err != nil {
		return "", err
	}

	for _, u := range users {
		c.userCache[u.Name] = u
	}
	for _, u := range users {
		if _, ok := c.userCache[u.Profile.DisplayName]; !ok {
			c.userCache[u.Profile.DisplayName] = u
		}
	}

	uc, ok := c.userCache[name]
	if ok {
		return fmt.Sprintf("<@%s>", uc.ID), nil
	}

	groups, err := c.client.GetUserGroupsContext(ctx)
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		c.userGroupCache[g.Handle] = g
	}
	gc, ok := c.userGroupCache[name]
	if ok {
		return fmt.Sprintf("<!subteam^%s>", gc.ID), nil
	}

	return fmt.Sprintf("<@%s|not found user or usergroup>", name), nil
}

func (c *Client) replaceBlockMentions(ctx context.Context, b slack.Block) error {
	switch v := b.(type) {
	case *slack.HeaderBlock:
		if v.Text != nil {
			v.Text.Text = c.ReplaceMentionsToMentionLinks(ctx, v.Text.Text)
		}
	case *slack.SectionBlock:
		if v.Text != nil {
			v.Text.Text = c.ReplaceMentionsToMentionLinks(ctx, v.Text.Text)
		}
		for _, f := range v.Fields {
			f.Text = c.ReplaceMentionsToMentionLinks(ctx, f.Text)
		}
	}
	return nil
}
