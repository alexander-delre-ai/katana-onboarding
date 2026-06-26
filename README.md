# katana-onboarding

Tooling for onboarding Komatsu (NA) engineers onto the Katana program. The repo contains two layers: a set of Python CLI scripts used for one-off operations, and a Go Slack bot deployed to Google Cloud Run that automates the same workflows interactively.

---

## Repository layout

```
katana_onboarding/     Python CLI scripts (original tooling)
katana-onboarder/      Go Slack bot + React frontend (Apps Platform app)
```

---

## `katana_onboarding` - Python CLI scripts

One-shot scripts for filing Jira IT service desk tickets, adding engineers to Slack channels, and creating AWS IAM users. Run locally as needed; they do not require a running server.

See [katana_onboarding/README.md](katana_onboarding/README.md) for full usage.

**Quick reference:**

```bash
# Provision Okta account + Slack channels for new engineers
python3 katana_onboarding/create_service_desk_request.py onboard \
  --user "First Last first.last@global.komatsu"

# Add engineers to additional Slack channels
python3 katana_onboarding/create_service_desk_request.py slack-add \
  --user "First Last first.last@global.komatsu" \
  --channels "#ext-katana-alerts"

# Create AWS IAM user and email credentials
python3 katana_onboarding/create_service_desk_request.py aws-user \
  --user "First Last first.last@global.komatsu"
```

**Setup:** Copy `katana_onboarding/constants.template.py` to `katana_onboarding/constants.py` and fill in your values. Set `ATLASSIAN_EMAIL` and `KATANA_ONBOARDER_API` (Atlassian API token) environment variables before running.

---

## `katana-onboarder` - Slack bot + web frontend

A Go service deployed on Apps Platform (Google Cloud Run) that exposes the same onboarding operations through a Slack bot and a React admin dashboard.

### What it does

**Slack bot:**
- `/katana-onboard new` - opens a modal to file an Okta + Slack provisioning ticket (Jira request type 244)
- `/katana-onboard slack-add` - opens a modal to add engineers to additional Slack channels (Jira request type 307)
- Mention shorthand: `@katana-onboarder onboard First Last email@global.komatsu`
- Mention shorthand with extra channels: `@katana-onboarder onboard First Last email to #channel1, #channel2`
- DM responses and a `/feedback` slash command for collecting feedback

**Web dashboard** (React + Tailwind, served from the same binary):
- Event log: recent bot activity
- Feedback submissions collected via the `/feedback` command
- Komatsu engineer roster with AWS account status (data source not yet connected)
- Pending access requests: open onboarding tickets from Jira (data source not yet connected)

### Stack

| Layer | Tech |
|-------|------|
| Backend | Go, Gin, `go.apps.applied.dev/lib/slacklib` |
| Frontend | React, Vite, Tailwind CSS |
| Hosting | Google Cloud Run via Apps Platform |
| Secrets | Google Secret Manager (`katana-onboarder-atlassian-email`, `katana-onboarder-atlassian-token`) |

### Local development

```bash
cd katana-onboarder

# Install dependencies
make deps

# Run backend and frontend concurrently (backend :8080, frontend :3000)
ENV=dev ATLASSIAN_EMAIL=you@applied.co ATLASSIAN_TOKEN=<token> make run

# Or run them separately
make backend    # Go server on :8080
make frontend   # Vite dev server on :3000
```

### Deploy

```bash
cd katana-onboarder
make deploy     # builds frontend, then: apps-platform app deploy --no-build
```

Credentials are read from Secret Manager in production. Locally, set `ENV=dev` and use environment variables instead.

### Wiring up incomplete data sources

Two API endpoints return `"configured": false` and empty lists because their upstream connections are not yet implemented:

- `GET /api/komatsu-users` - should list the AWS `katana` IAM group members in account `055333015386` (one `ListUsersForGroup("katana")` call; `first.last` username presence indicates an active AWS account)
- `GET /api/pending-access` - should run a JQL search of Jira Service Desk 31 for open request types 244 and 307 using the existing Atlassian credentials
