# AWS
AWS_ACCOUNT_ID = "055333015386"
AWS_IAM_GROUP = "katana"
AWS_REGION = "us-east-1"
AWS_CONSOLE_URL = f"https://{AWS_ACCOUNT_ID}.signin.aws.amazon.com/console"

# Atlassian
BASE_URL = "https://appliedintuition.atlassian.net"
SERVICE_DESK_ID = "31"
REQUEST_TYPE_ONBOARD = "244"  # Get IT help
REQUEST_TYPE_SLACK = "307"    # Slack Help

LOCATIONS: dict[str, str] = {
    "all":          "13271",
    "ann-arbor":    "12208",
    "australia":    "34481",
    "episci":       "27986",
    "fort-walton":  "23773",
    "india":        "28019",
    "london":       "25968",
    "munich":       "12211",
    "remote":       "16668",
    "rockwell":     "26831",
    "san-diego":    "16408",
    "seoul":        "12230",
    "sf-depot":     "16474",
    "stockholm":    "12232",
    "stuttgart":    "15210",
    "svl-460w":     "16706",
    "svl-490w":     "16707",
    "svl-860w":     "23917",
    "tokyo":        "12209",
    "washington":   "12231",
}

PRIORITIES: dict[str, str] = {
    "P0": "1",
    "P1": "2",
    "P2": "3",
    "P3": "4",
    "P4": "5",
}

DEFAULT_LOCATION = "svl-860w"
DEFAULT_PRIORITY = "P1"

# onboard defaults
ONBOARD_DEFAULT_SUMMARY = "Add Komatsu user to ext-komatsu okta group and slack channels"
ONBOARD_DESCRIPTION_TEMPLATE = (
    "Hello, Please provision below users with ext-komatsu.applied.co accounts "
    "under the ext-komatsu okta group.\n\n"
    "{user_lines}\n\n"
    "Please also add the komatsu user to the below slack channels:\n\n"
    "[#ext-program-katana|https://grid-appliedint.enterprise.slack.com/archives/C098TS704KG]\n"
    "[#ext-program-katana-toolchain|https://grid-appliedint.enterprise.slack.com/archives/C09LSNE9Q0Y]"
)

# slack-add defaults
SLACK_DEFAULT_SUMMARY = "Add Komatsu user to Slack channels"
SLACK_DESCRIPTION_TEMPLATE = (
    "Hello, Please add the below users to the following Slack channels:\n\n"
    "{user_lines}\n\n"
    "Channels:\n\n"
    "{channel_lines}"
)
