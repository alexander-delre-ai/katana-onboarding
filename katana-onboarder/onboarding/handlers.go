// Package onboarding implements the /katana-onboard Slack command, which files
// Komatsu engineer onboarding requests in Jira Service Management. It is the Go
// port of scripts/katana_onboarding (the onboard and slack-add subcommands).
package onboarding

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"slack-go/eventlog"

	"github.com/slack-go/slack"
	"go.apps.applied.dev/lib/slacklib"
	"go.uber.org/zap"
)

// Modal callback IDs and form block/action IDs. ViewSubmission handlers and
// GetValue both key off these, so the modal builders and the handlers must agree.
const (
	onboardModalCallback  = "katana_onboard_new"
	slackAddModalCallback = "katana_onboard_slack_add"

	blockEngineers  = "engineers_block"
	actionEngineers = "engineers_input"
	blockLocation   = "location_block"
	actionLocation  = "location_select"
	blockPriority   = "priority_block"
	actionPriority  = "priority_select"
	blockChannels   = "channels_block"
	actionChannels  = "channels_input"
)

// Register loads the Atlassian credentials and wires up the /katana-onboard
// command and its modal submission handlers. It returns an error if the
// credentials cannot be loaded, in which case no handlers are registered.
func Register(ctx context.Context, bot *slacklib.Bot) error {
	c, err := loadCreds(ctx)
	if err != nil {
		return err
	}
	registerHandlers(bot, c)
	return nil
}

// RegisterUnconfigured registers stub handlers that tell users why the bot is
// unavailable, so they get a clear message instead of silence.
func RegisterUnconfigured(bot *slacklib.Bot, reason error) {
	msg := fmt.Sprintf(":warning: Katana onboarding is unavailable: credentials could not be loaded (%s). Contact the bot owner to check Secret Manager configuration.", reason)

	bot.Command("/katana-onboard", func(ctx *slacklib.CommandContext) {
		ctx.Reply(msg)
	})

	bot.OnMention(func(ctx *slacklib.MentionContext) {
		threadTS := ctx.ThreadTS
		if threadTS == "" {
			threadTS = ctx.EventTS
		}
		if _, err := bot.SendMessageInThread(context.Background(), ctx.ChannelID, msg, threadTS); err != nil {
			zap.L().Error("failed to reply to mention (unconfigured)", zap.Error(err))
		}
	})
}

func registerHandlers(bot *slacklib.Bot, c creds) {
	bot.Command("/katana-onboard", func(ctx *slacklib.CommandContext) {
		sub := ""
		if fields := strings.Fields(ctx.Text); len(fields) > 0 {
			sub = strings.ToLower(fields[0])
		}
		eventlog.Add("slash_command", ctx.UserID, ctx.ChannelID, "/katana-onboard "+sub)

		switch sub {
		case "new":
			if err := openView(bot, ctx.Context(), ctx.TriggerID, buildOnboardModal()); err != nil {
				zap.L().Error("failed to open onboard modal", zap.Error(err))
			}
		case "slack-add":
			if err := openView(bot, ctx.Context(), ctx.TriggerID, buildSlackAddModal()); err != nil {
				zap.L().Error("failed to open slack-add modal", zap.Error(err))
			}
		default:
			ctx.Reply("*Katana onboarding*\n" +
				"• `/katana-onboard new` - file the Okta + Slack provisioning ticket for new engineers\n" +
				"• `/katana-onboard slack-add` - add engineers to additional Slack channels")
		}
	})

	bot.ViewSubmission(onboardModalCallback, func(ctx *slacklib.ViewContext) {
		userID := ctx.UserID
		engineers, err := parseEngineers(ctx.GetValue(blockEngineers, actionEngineers))
		if err != nil {
			notify(bot, userID, "Could not file the onboarding ticket: "+err.Error())
			return
		}
		location := ctx.GetValue(blockLocation, actionLocation)
		priority := ctx.GetValue(blockPriority, actionPriority)
		extraChannels := parseChannels(ctx.GetValue(blockChannels, actionChannels))
		eventlog.Add("katana_onboard", userID, "", fmt.Sprintf("%d engineer(s), %s/%s", len(engineers), location, priority))
		notify(bot, userID, fmt.Sprintf(":hourglass_flowing_sand: Filing onboarding ticket for %d engineer(s)...", len(engineers)))

		go func() {
			link, err := createOnboardRequest(c, engineers, extraChannels, location, priority)
			if err != nil {
				zap.L().Error("onboard request failed", zap.Error(err))
				notify(bot, userID, ":x: Onboarding ticket failed: "+err.Error())
				return
			}
			notify(bot, userID, fmt.Sprintf(":white_check_mark: Onboarding ticket created for %d engineer(s): %s", len(engineers), link))
		}()
	})

	bot.ViewSubmission(slackAddModalCallback, func(ctx *slacklib.ViewContext) {
		userID := ctx.UserID
		engineers, err := parseEngineers(ctx.GetValue(blockEngineers, actionEngineers))
		if err != nil {
			notify(bot, userID, "Could not file the Slack channel request: "+err.Error())
			return
		}
		channels := parseChannels(ctx.GetValue(blockChannels, actionChannels))
		if len(channels) == 0 {
			notify(bot, userID, "Could not file the Slack channel request: no channels provided")
			return
		}
		eventlog.Add("katana_slack_add", userID, "", fmt.Sprintf("%d engineer(s) -> %s", len(engineers), strings.Join(channels, ", ")))
		notify(bot, userID, fmt.Sprintf(":hourglass_flowing_sand: Filing Slack channel request for %d engineer(s) to %s...", len(engineers), strings.Join(channels, ", ")))

		go func() {
			link, err := createSlackAddRequest(c, engineers, channels, defaultLocation, defaultPriority)
			if err != nil {
				zap.L().Error("slack-add request failed", zap.Error(err))
				notify(bot, userID, ":x: Slack channel request failed: "+err.Error())
				return
			}
			notify(bot, userID, fmt.Sprintf(":white_check_mark: Slack channel request created for %d engineer(s): %s", len(engineers), link))
		}()
	})

	// Mention shorthand, e.g. "@katana-onboarder onboard First Last email@global.komatsu".
	// This overrides the default mention handler whenever onboarding is configured.
	bot.OnMention(func(ctx *slacklib.MentionContext) {
		text := strings.TrimSpace(unwrapSlackEntities(ctx.Text))
		eventlog.Add("app_mention", ctx.UserID, ctx.ChannelID, text)

		threadTS := ctx.ThreadTS
		if threadTS == "" {
			threadTS = ctx.EventTS
		}
		reply := func(msg string) {
			if _, err := bot.SendMessageInThread(context.Background(), ctx.ChannelID, msg, threadTS); err != nil {
				zap.L().Error("failed to reply to mention", zap.Error(err))
			}
		}

		fields := strings.Fields(text)
		if len(fields) == 0 {
			reply(mentionHelp)
			return
		}
		sub := strings.ToLower(fields[0])
		rest := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))

		switch sub {
		case "onboard":
			usersRaw := rest
			var extraChannels []string
			if u, channelsRaw, ok := splitOnTo(rest); ok {
				usersRaw = u
				extraChannels = parseChannels(channelsRaw)
			}
			engineers, err := parseEngineers(usersRaw)
			if err != nil {
				reply("Could not file the onboarding ticket: " + err.Error())
				return
			}
			reply(fmt.Sprintf(":hourglass_flowing_sand: Filing onboarding ticket for %d engineer(s)...", len(engineers)))
			go func() {
				link, err := createOnboardRequest(c, engineers, extraChannels, defaultLocation, defaultPriority)
				if err != nil {
					zap.L().Error("onboard request failed", zap.Error(err))
					reply(":x: Onboarding ticket failed: " + err.Error())
					return
				}
				reply(fmt.Sprintf(":white_check_mark: Onboarding ticket created for %d engineer(s): %s", len(engineers), link))
			}()
		case "slack-add":
			users, channelsRaw, ok := splitOnTo(rest)
			if !ok {
				reply("Usage: `@katana-onboarder slack-add First Last email@global.komatsu to #channel1, #channel2`")
				return
			}
			engineers, err := parseEngineers(users)
			if err != nil {
				reply("Could not file the Slack channel request: " + err.Error())
				return
			}
			channels := parseChannels(channelsRaw)
			if len(channels) == 0 {
				reply("Could not file the Slack channel request: no channels provided")
				return
			}
			reply(fmt.Sprintf(":hourglass_flowing_sand: Filing Slack channel request for %d engineer(s) to %s...", len(engineers), strings.Join(channels, ", ")))
			go func() {
				link, err := createSlackAddRequest(c, engineers, channels, defaultLocation, defaultPriority)
				if err != nil {
					zap.L().Error("slack-add request failed", zap.Error(err))
					reply(":x: Slack channel request failed: " + err.Error())
					return
				}
				reply(fmt.Sprintf(":white_check_mark: Slack channel request created for %d engineer(s): %s", len(engineers), link))
			}()
		default:
			reply(mentionHelp)
		}
	})
}

const mentionHelp = "*Katana onboarding* - mention me with:\n" +
	"• `onboard First Last email@global.komatsu` (one or more, comma-separated)\n" +
	"• `onboard First Last email@global.komatsu to #channel1, #channel2` (extra channels beyond defaults)\n" +
	"• `slack-add First Last email@global.komatsu to #channel1, #channel2`\n" +
	"Tickets use the default location (svl-860w) and priority (P1)."

var (
	reSlackUser    = regexp.MustCompile(`<@[^>]+>`)
	reSlackMailto  = regexp.MustCompile(`<mailto:([^|>]+)(?:\|[^>]*)?>`)
	reSlackChannel = regexp.MustCompile(`<#[^|>]+\|([^>]+)>`)
	reSlackLink    = regexp.MustCompile(`<(https?://[^|>]+)(?:\|[^>]*)?>`)
)

// unwrapSlackEntities normalizes Slack's auto-linked entities in message text:
// it drops user mentions (e.g. the leading "<@BOTID>"), and converts
// <mailto:a@b|a@b> -> a@b, <#C123|name> -> #name, <http://x|x> -> http://x.
func unwrapSlackEntities(s string) string {
	s = reSlackUser.ReplaceAllString(s, "")
	s = reSlackMailto.ReplaceAllString(s, "$1")
	s = reSlackChannel.ReplaceAllString(s, "#$1")
	s = reSlackLink.ReplaceAllString(s, "$1")
	return s
}

// splitOnTo splits "users... to channels..." on the first " to " separator.
func splitOnTo(s string) (users, channels string, ok bool) {
	idx := strings.Index(strings.ToLower(s), " to ")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+4:]), true
}

// notify DMs the user with a result. Uses context.Background() because it runs
// in a goroutine after the request that triggered it has returned.
func notify(bot *slacklib.Bot, userID, text string) {
	if _, err := bot.SendDM(context.Background(), userID, text); err != nil {
		zap.L().Error("failed to send result DM", zap.Error(err))
	}
}

func parseChannels(raw string) []string {
	var channels []string
	for _, c := range strings.Split(raw, ",") {
		if c = strings.TrimSpace(c); c != "" {
			channels = append(channels, c)
		}
	}
	return channels
}

// openView opens a modal built as a raw slack.ModalViewRequest. slacklib's
// modal builder does not expose initial values, so we build views directly to
// support prefilled fields and default-selected options.
func openView(bot *slacklib.Bot, ctx context.Context, triggerID string, view slack.ModalViewRequest) error {
	client, err := bot.Client()
	if err != nil {
		return err
	}
	_, err = client.OpenViewContext(ctx, triggerID, view)
	return err
}
