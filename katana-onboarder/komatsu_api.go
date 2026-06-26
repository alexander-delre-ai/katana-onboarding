package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// komatsuUser is a row in the engineers roster. HasAWSAccount is populated from
// the AWS katana IAM group once that data source is wired up.
type komatsuUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	HasAWSAccount bool   `json:"has_aws_account"`
}

// handleGetKomatsuUsers returns the ext-komatsu engineer roster with AWS account
// status. The data source (AWS katana IAM group) is not connected yet, so this
// returns configured=false with an empty list. To wire it up: read-only IAM
// ListUsersForGroup("katana") in account 055333015386, mapping first.last
// usernames; the presence in that group is the "has AWS account" signal.
func handleGetKomatsuUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"configured": false,
			"source":     "AWS katana IAM group",
			"users":      []komatsuUser{},
		})
	}
}

// pendingAccessRequest is an open onboarding (terminal) ticket awaiting IT.
type pendingAccessRequest struct {
	Engineer  string `json:"engineer"`
	Email     string `json:"email"`
	TicketKey string `json:"ticket_key"`
	TicketURL string `json:"ticket_url"`
	Status    string `json:"status"`
	Created   string `json:"created"`
}

// handleGetPendingAccess returns engineers whose onboarding (terminal) tickets
// are still open, each linking to its Jira ticket. Not connected yet; to wire
// it up: a JQL search of service desk 31 for open request types 244/307 using
// the existing Atlassian credentials, returning each issue's browse URL.
func handleGetPendingAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"configured": false,
			"source":     "Jira Service Desk (open onboarding tickets)",
			"requests":   []pendingAccessRequest{},
		})
	}
}
