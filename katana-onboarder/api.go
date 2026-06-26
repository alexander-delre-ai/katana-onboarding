package main

import (
	"net/http"

	"slack-go/eventlog"

	"go.apps.applied.dev/lib/slacklib"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// errorHandler is middleware that handles errors added via c.Error()
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			zap.L().Error("api error: %v", zap.Error(err), zap.String("path", c.Request.URL.Path))

			status := c.Writer.Status()
			if status == http.StatusOK {
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"success": false, "error": err.Error()})
		}
	}
}

func registerAPIRoutes(r *gin.Engine, bot *slacklib.Bot) {
	api := r.Group("/api")
	api.Use(errorHandler())

	api.POST("/send-message", handleSendMessage(bot))
	api.POST("/send-message-with-button", handleSendMessageWithButton(bot))
	api.POST("/send-dm", handleSendDM(bot))
	api.GET("/members", handleGetMembers(bot))
	api.GET("/user", handleGetUser(bot))
	api.GET("/events", handleGetEvents())
	api.GET("/feedback", handleGetFeedback())
	api.GET("/komatsu-users", handleGetKomatsuUsers())
	api.GET("/pending-access", handleGetPendingAccess())
}

type sendMessageRequest struct {
	Channel  string `json:"channel" binding:"required"`
	Text     string `json:"text" binding:"required"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

func handleSendMessage(bot *slacklib.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var result *slacklib.MessageResult
		var err error
		if req.ThreadTS != "" {
			result, err = bot.SendMessageInThread(c, req.Channel, req.Text, req.ThreadTS)
		} else {
			result, err = bot.SendMessage(c, req.Channel, req.Text)
		}
		if err != nil {
			c.Error(err)
			return
		}

		eventlog.Add("message_sent", "", req.Channel, req.Text)
		c.JSON(http.StatusOK, gin.H{"success": true, "channel": result.ChannelID, "timestamp": result.Timestamp})
	}
}

type sendMessageWithButtonRequest struct {
	Channel    string `json:"channel" binding:"required"`
	Text       string `json:"text" binding:"required"`
	ButtonText string `json:"button_text"`
	ActionID   string `json:"action_id"`
}

func handleSendMessageWithButton(bot *slacklib.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendMessageWithButtonRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		buttonText, actionID := req.ButtonText, req.ActionID
		if buttonText == "" {
			buttonText = "Click"
		}
		if actionID == "" {
			actionID = "button_click"
		}

		blocks := []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, req.Text, false, false), nil, nil),
			slack.NewActionBlock("", slack.NewButtonBlockElement(actionID, "click", slack.NewTextBlockObject(slack.PlainTextType, buttonText, false, false))),
		}

		result, err := bot.SendMessageWithBlocks(c, req.Channel, blocks)
		if err != nil {
			c.Error(err)
			return
		}

		eventlog.Add("message_sent", "", req.Channel, req.Text+" [with button]")
		c.JSON(http.StatusOK, gin.H{"success": true, "channel": result.ChannelID, "timestamp": result.Timestamp})
	}
}

type sendDMRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

func handleSendDM(bot *slacklib.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendDMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		result, err := bot.SendDM(c, req.UserID, req.Text)
		if err != nil {
			c.Error(err)
			return
		}

		eventlog.Add("dm_sent", req.UserID, result.ChannelID, req.Text)
		c.JSON(http.StatusOK, gin.H{"success": true, "channel": result.ChannelID})
	}
}

type getMembersRequest struct {
	Channel string `form:"channel" binding:"required"`
}

func handleGetMembers(bot *slacklib.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req getMembersRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		members, err := bot.GetChannelMembers(c, req.Channel)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "members": members})
	}
}

type getUserRequest struct {
	UserID string `form:"user_id" binding:"required"`
}

func handleGetUser(bot *slacklib.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req getUserRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		user, err := bot.GetUserInfo(c, req.UserID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user": gin.H{
				"id":        user.ID,
				"name":      user.Name,
				"real_name": user.RealName,
				"email":     user.Profile.Email,
				"title":     user.Profile.Title,
				"image":     user.Profile.Image72,
			},
		})
	}
}

func handleGetEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		events := eventlog.GetRecent()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"events":  events,
		})
	}
}

func handleGetFeedback() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Feedback is stored in eventlog as "feedback_submitted" events
		// Filter those out and return them
		allEvents := eventlog.GetRecent()
		var submissions []gin.H

		for _, e := range allEvents {
			if e.Type == "feedback_submitted" {
				submissions = append(submissions, gin.H{
					"id":           e.ID,
					"user_id":      e.User,
					"description":  e.Text,
					"category":     "feedback",
					"urgency":      "medium",
					"submitted_at": e.Timestamp,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"submissions": submissions,
		})
	}
}
