# Komatsu Onboarding Scripts

Automation for onboarding new Komatsu engineers onto the Katana program. Covers Jira IT requests, Slack channel provisioning, and AWS IAM user creation.

---

## Setup

### 1. Copy and fill in constants

```bash
cp scripts/katana_onboarding/constants.template.py scripts/katana_onboarding/constants.py
```

Edit `constants.py` with your organization's values (AWS account ID, Atlassian service desk IDs, etc.). This file is gitignored and never committed.

### 2. Set environment variables

| Variable | Used by | Description |
|----------|---------|-------------|
| `ATLASSIAN_EMAIL` | `onboard`, `slack-add` | Your Atlassian account email |
| `KATANA_ONBOARDER_API` | `onboard`, `slack-add` | Atlassian API token from [id.atlassian.com](https://id.atlassian.com) |
| `SES_SENDER_EMAIL` | `aws-user` | Verified SES sender address (default: `noreply@applied.co`) |

AWS credentials follow the standard boto3 chain (`~/.aws/credentials`, env vars, or instance profile).

---

## Entry Point

All subcommands are run through a single script:

```bash
python3 scripts/katana_onboarding/create_service_desk_request.py <subcommand> [options]
```

---

## Subcommands

### `onboard` - Provision Okta + Slack for new engineers

Files a **Get IT help** request (type 244) in the [INT] IT Operations (CSI) service desk to provision `ext-komatsu.applied.co` Okta accounts and add engineers to the standard Slack channels.

**Defaults:**
- Service desk: `[INT] IT Operations (CSI)` (ID 31)
- Request type: Get IT help (244)
- Location: SVL-860W
- Priority: P1
- Slack channels included: `#ext-program-katana`, `#ext-program-katana-toolchain`

**Usage:**

```bash
# Single engineer
python3 scripts/katana_onboarding/create_service_desk_request.py onboard \
  --user "First Last first.last@global.komatsu"

# Multiple engineers
python3 scripts/katana_onboarding/create_service_desk_request.py onboard \
  --user "First Last first.last@global.komatsu, Second Engineer second.engineer@global.komatsu"

# From a CSV file
python3 scripts/katana_onboarding/create_service_desk_request.py onboard \
  --users-file engineers.csv

# Override location or priority
python3 scripts/katana_onboarding/create_service_desk_request.py onboard \
  --user "First Last first.last@global.komatsu" \
  --location tokyo --priority P2

# Include additional Slack channels beyond the defaults
python3 scripts/katana_onboarding/create_service_desk_request.py onboard \
  --user "First Last first.last@global.komatsu" \
  --channels "#ext-katana-alerts, #general"
```

**Output:** URL of the created Jira ticket.

**Notes:**
- Engineer emails must end with `@global.komatsu` (validated before submission)
- One ticket is created per run, listing all engineers in the description

---

### `slack-add` - Add engineers to additional Slack channels

Files a **Slack Help** request (type 307) to add engineers to channels beyond the onboarding defaults.

**Usage:**

```bash
python3 scripts/katana_onboarding/create_service_desk_request.py slack-add \
  --user "First Last first.last@global.komatsu" \
  --channels "#ext-program-katana, #another-channel"
```

**Output:** URL of the created Jira ticket.

---

### `aws-user` - Create AWS IAM user and email credentials

Creates an IAM user with username `first.last`, adds them to the `katana` IAM group, generates a temporary password (reset required on first login) and access keys, then emails the credentials via SES.

**Usage:**

```bash
# Credentials sent to the engineer's own email
python3 scripts/katana_onboarding/create_service_desk_request.py aws-user \
  --user "First Last first.last@global.komatsu"

# Send credentials to a different address
python3 scripts/katana_onboarding/create_service_desk_request.py aws-user \
  --user "First Last first.last@global.komatsu" \
  --send-to manager@applied.co

# Multiple engineers from a CSV
python3 scripts/katana_onboarding/create_service_desk_request.py aws-user \
  --users-file engineers.csv
```

**What it does per engineer:**
1. Creates IAM user `first.last` in account `055333015386`
2. Adds user to the `katana` IAM group
3. Creates a login profile with a temporary password (`PasswordResetRequired=True`)
4. Emails console URL, username, and temp password via SES

**Note:** The `aws-user` subcommand does not validate the `@global.komatsu` domain since the email is used as the credentials recipient, not an Okta identity.

---

## CSV Format

For `--users-file`, provide a CSV with columns `first_name`, `last_name`, `email`. The header row is optional:

```csv
first_name,last_name,email
First,Last,first.last@global.komatsu
Second,Engineer,second.engineer@global.komatsu
```

```csv
First,Last,first.last@global.komatsu
Second,Engineer,second.engineer@global.komatsu
```

---

## File Structure

```
scripts/katana_onboarding/
  create_service_desk_request.py   # CLI entry point - imports from both modules below
  terminal_requests.py             # Jira/Atlassian logic: onboard, slack-add, Engineer model
  aws_users.py                     # AWS IAM + SES email logic
  constants.py                     # Local config (gitignored - never commit)
  constants.template.py            # Committed template - copy to constants.py to get started
  README.md
```

---

## Common Options

All subcommands accept `--user` and `--users-file`. Jira subcommands additionally accept:

| Flag | Description |
|------|-------------|
| `--location` | Office location slug (see `constants.py` for valid values) |
| `--priority` | `P0` through `P4` |
| `--summary` | Override the default ticket summary |
| `--description` | Override the default ticket description |
| `--participants` | Jira account IDs to add as request participants |
| `--channels` | (`onboard`, `slack-add`) Comma-separated Slack channels. For `onboard`, appended to the two defaults; for `slack-add`, required and used as the full channel list |
