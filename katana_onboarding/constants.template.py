# Copy this file to constants.py and fill in the values before running.
#   cp scripts/constants.template.py scripts/constants.py

# AWS
AWS_ACCOUNT_ID = "REPLACE_ME"           # e.g. "123456789012"
AWS_IAM_GROUP = "REPLACE_ME"            # e.g. "katana"
AWS_REGION = "REPLACE_ME"               # e.g. "us-east-1"
AWS_CONSOLE_URL = f"https://{AWS_ACCOUNT_ID}.signin.aws.amazon.com/console"

# Atlassian
BASE_URL = "REPLACE_ME"                 # e.g. "https://yourorg.atlassian.net"
SERVICE_DESK_ID = "REPLACE_ME"          # numeric string, find via --list-desks
REQUEST_TYPE_ONBOARD = "REPLACE_ME"     # numeric string for "Get IT help"
REQUEST_TYPE_SLACK = "REPLACE_ME"       # numeric string for "Slack Help"

LOCATIONS: dict[str, str] = {
    # "slug": "numeric-id"
    # Populate from /rest/servicedeskapi/servicedesk/{id}/requesttype/{type}/field
    "REPLACE_ME": "REPLACE_ME",
}

PRIORITIES: dict[str, str] = {
    # "label": "numeric-id"
    # Populate from the priority field validValues for your request type
    "P0": "REPLACE_ME",
    "P1": "REPLACE_ME",
    "P2": "REPLACE_ME",
    "P3": "REPLACE_ME",
    "P4": "REPLACE_ME",
}

DEFAULT_LOCATION = "REPLACE_ME"         # must be a key in LOCATIONS
DEFAULT_PRIORITY = "REPLACE_ME"         # must be a key in PRIORITIES

# onboard defaults
ONBOARD_DEFAULT_SUMMARY = "REPLACE_ME"
ONBOARD_DESCRIPTION_TEMPLATE = (
    "REPLACE_ME\n\n"
    "{user_lines}\n\n"
    "REPLACE_ME"
)

# slack-add defaults
SLACK_DEFAULT_SUMMARY = "REPLACE_ME"
SLACK_DESCRIPTION_TEMPLATE = (
    "REPLACE_ME\n\n"
    "{user_lines}\n\n"
    "Channels:\n\n"
    "{channel_lines}"
)
