package onboarding

import (
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Atlassian / Jira Service Desk configuration.
// Ported from scripts/katana_onboarding/constants.py.
const (
	jiraBaseURL         = "https://appliedintuition.atlassian.net"
	serviceDeskID       = "31"
	requestTypeOnboard  = "244" // Get IT help
	requestTypeSlack    = "307" // Slack Help
	locationCustomField = "customfield_10513"

	defaultLocation = "svl-860w"
	defaultPriority = "P1"

	onboardDefaultSummary = "Add Komatsu user to ext-komatsu okta group and slack channels"
	slackDefaultSummary   = "Add Komatsu user to Slack channels"

	// Prefilled into the slack-add modal's channels field.
	defaultSlackChannels = "#ext-program-katana, #ext-program-katana-toolchain"
)

const onboardDescriptionTemplate = "Hello, Please provision below users with ext-komatsu.applied.co accounts " +
	"under the ext-komatsu okta group.\n\n" +
	"%s\n\n" +
	"Please also add the komatsu user to the below slack channels:\n\n" +
	"[#ext-program-katana|https://grid-appliedint.enterprise.slack.com/archives/C098TS704KG]\n" +
	"[#ext-program-katana-toolchain|https://grid-appliedint.enterprise.slack.com/archives/C09LSNE9Q0Y]"

const slackDescriptionTemplate = "Hello, Please add the below users to the following Slack channels:\n\n" +
	"%s\n\n" +
	"Channels:\n\n" +
	"%s"

// orderedKV keeps select options in a stable, human-friendly order while
// mapping a display key to its Jira field id.
type orderedKV struct {
	Key string
	ID  string
}

// locations mirrors constants.LOCATIONS (slug -> Jira field id).
var locations = []orderedKV{
	{"all", "13271"},
	{"ann-arbor", "12208"},
	{"australia", "34481"},
	{"episci", "27986"},
	{"fort-walton", "23773"},
	{"india", "28019"},
	{"london", "25968"},
	{"munich", "12211"},
	{"remote", "16668"},
	{"rockwell", "26831"},
	{"san-diego", "16408"},
	{"seoul", "12230"},
	{"sf-depot", "16474"},
	{"stockholm", "12232"},
	{"stuttgart", "15210"},
	{"svl-460w", "16706"},
	{"svl-490w", "16707"},
	{"svl-860w", "23917"},
	{"tokyo", "12209"},
	{"washington", "12231"},
}

// priorities mirrors constants.PRIORITIES (label -> Jira priority id).
var priorities = []orderedKV{
	{"P0", "1"},
	{"P1", "2"},
	{"P2", "3"},
	{"P3", "4"},
	{"P4", "5"},
}

func lookupID(opts []orderedKV, key string) (string, bool) {
	for _, o := range opts {
		if o.Key == key {
			return o.ID, true
		}
	}
	return "", false
}

// creds holds the Service Desk API credentials.
type creds struct {
	Email string
	Token string
}

// loadCreds reads the Atlassian email + API token. Locally (ENV=dev) it reads
// them from environment variables; in production it reads them from Secret
// Manager. Call this once at startup and cache the result.
func loadCreds(ctx context.Context) (creds, error) {
	if os.Getenv("ENV") == "dev" {
		c := creds{
			Email: os.Getenv("ATLASSIAN_EMAIL"),
			Token: os.Getenv("ATLASSIAN_TOKEN"),
		}
		if c.Email == "" || c.Token == "" {
			return c, fmt.Errorf("ENV=dev: set ATLASSIAN_EMAIL and ATLASSIAN_TOKEN env vars")
		}
		return c, nil
	}

	email, err := getSecret(ctx, "katana-onboarder-atlassian-email")
	if err != nil {
		return creds{}, err
	}
	token, err := getSecret(ctx, "katana-onboarder-atlassian-token")
	if err != nil {
		return creds{}, err
	}
	return creds{Email: email, Token: token}, nil
}

func getSecret(ctx context.Context, name string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("secret manager client: %w", err)
	}
	defer client.Close()

	path := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", os.Getenv("PROJECT_ID"), name)
	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: path})
	if err != nil {
		return "", fmt.Errorf("access secret %s: %w", name, err)
	}
	return string(result.Payload.Data), nil
}
