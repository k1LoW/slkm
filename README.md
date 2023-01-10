# slkm

slkm is `github.com/slack-go/slack` wrapper package for posting message.

This package converts channels and mentions of users and user groups into the correct ID notation.

- `#channel-name` -> `CXXXXXXXX`
- `@username` -> `<@UXXXXXXXX>`
- `@usergroup-name` -> `<!subteam^SXXXXXXXX>`

## Usage

``` go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/k1LoW/slkm"
	"github.com/slack-go/slack"
)

const (
	notifyChannel = "#service-alerts"
)

func main() {
	ctx := context.Background()
	c, err := slkm.New()
	if err != nil {
		log.Fatal(err)
	}
	c.SetUsername("wakeup-bot")
	blocks := []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "Wake up @k1low !!", false, false), nil, nil),
	}
	if err := c.PostMessage(ctx, notifyChannel, blocks...); err != nil {
		log.Fatal(err)
	}
}

```

## Required scope of `SLACK_API_TOKEN`

- `channel:read`
- `chat:write`
- `chat:write.public`
- `users:read`
- `usergroups:read`
- `chat:write.customize` ( optional )
