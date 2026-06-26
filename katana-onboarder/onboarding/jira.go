package onboarding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const komatsuEmailDomain = "@global.komatsu"

// Engineer is a single person to onboard. Mirrors the Python Engineer dataclass.
type Engineer struct {
	FirstName string
	LastName  string
	Email     string
}

// parseEngineers parses "First Last email" entries separated by newlines or
// commas (modal text areas use newlines; a comma-separated line also works).
// Every email must end with @global.komatsu, matching the Python validation.
func parseEngineers(raw string) ([]Engineer, error) {
	entries := strings.FieldsFunc(raw, func(r rune) bool { return r == '\n' || r == ',' })

	var engineers []Engineer
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.Fields(entry)
		if len(parts) != 3 {
			return nil, fmt.Errorf("each engineer must be 'First Last email', got %q", entry)
		}
		email := parts[2]
		if !strings.HasSuffix(email, komatsuEmailDomain) {
			return nil, fmt.Errorf("invalid email %q: must end with %s", email, komatsuEmailDomain)
		}
		engineers = append(engineers, Engineer{FirstName: parts[0], LastName: parts[1], Email: email})
	}
	if len(engineers) == 0 {
		return nil, fmt.Errorf("no engineers provided")
	}
	return engineers, nil
}

func userLines(engineers []Engineer) string {
	lines := make([]string, len(engineers))
	for i, e := range engineers {
		lines[i] = fmt.Sprintf("%s %s | %s", e.FirstName, e.LastName, e.Email)
	}
	return strings.Join(lines, "\n")
}

type serviceDeskPayload struct {
	ServiceDeskID      string         `json:"serviceDeskId"`
	RequestTypeID      string         `json:"requestTypeId"`
	RequestFieldValues map[string]any `json:"requestFieldValues"`
}

// postServiceDeskRequest files a Jira Service Management request and returns
// the web link of the created ticket.
func postServiceDeskRequest(c creds, requestTypeID, summary, description, locationID, priorityID string) (string, error) {
	payload := serviceDeskPayload{
		ServiceDeskID: serviceDeskID,
		RequestTypeID: requestTypeID,
		RequestFieldValues: map[string]any{
			"summary":           summary,
			"description":       description,
			locationCustomField: map[string]string{"id": locationID},
			"priority":          map[string]string{"id": priorityID},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, jiraBaseURL+"/rest/servicedeskapi/request", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	auth := base64.StdEncoding.EncodeToString([]byte(c.Email + ":" + c.Token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-ExperimentalApi", "opt-in")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call jira: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("jira returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		IssueKey string `json:"issueKey"`
		Links    struct {
			Web string `json:"web"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode jira response: %w", err)
	}

	switch {
	case result.Links.Web != "":
		return result.Links.Web, nil
	case result.IssueKey != "":
		return jiraBaseURL + "/browse/" + result.IssueKey, nil
	default:
		return strings.TrimSpace(string(respBody)), nil
	}
}

// createOnboardRequest files the Okta + Slack provisioning ticket (type 244).
// extraChannels are appended to the two default channels already in the template.
func createOnboardRequest(c creds, engineers []Engineer, extraChannels []string, locationSlug, priorityLabel string) (string, error) {
	locationID, ok := lookupID(locations, locationSlug)
	if !ok {
		return "", fmt.Errorf("unknown location %q", locationSlug)
	}
	priorityID, ok := lookupID(priorities, priorityLabel)
	if !ok {
		return "", fmt.Errorf("unknown priority %q", priorityLabel)
	}
	description := fmt.Sprintf(onboardDescriptionTemplate, userLines(engineers))
	if len(extraChannels) > 0 {
		description += "\n" + strings.Join(extraChannels, "\n")
	}
	return postServiceDeskRequest(c, requestTypeOnboard, onboardDefaultSummary, description, locationID, priorityID)
}

// createSlackAddRequest files a ticket to add engineers to channels (type 307).
func createSlackAddRequest(c creds, engineers []Engineer, channels []string, locationSlug, priorityLabel string) (string, error) {
	locationID, ok := lookupID(locations, locationSlug)
	if !ok {
		return "", fmt.Errorf("unknown location %q", locationSlug)
	}
	priorityID, ok := lookupID(priorities, priorityLabel)
	if !ok {
		return "", fmt.Errorf("unknown priority %q", priorityLabel)
	}
	description := fmt.Sprintf(slackDescriptionTemplate, userLines(engineers), strings.Join(channels, "\n"))
	return postServiceDeskRequest(c, requestTypeSlack, slackDefaultSummary, description, locationID, priorityID)
}
