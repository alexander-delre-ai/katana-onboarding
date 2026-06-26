#!/usr/bin/env python3
"""
Komatsu onboarding CLI.

Subcommands:
  onboard    Jira: Okta + Slack provisioning for new engineers (type 244)
  slack-add  Jira: Add engineers to additional Slack channels (type 307)
  aws-user   AWS:  Create IAM user (first.last), add to katana group, email credentials

Auth:
  Atlassian  --requestor-email / ATLASSIAN_EMAIL + KATANA_ONBOARDER_API
  AWS        standard boto3 credential chain
  Email      SES_SENDER_EMAIL env var (default: noreply@applied.co)
"""

import os
import argparse

from constants import LOCATIONS, PRIORITIES, DEFAULT_LOCATION, DEFAULT_PRIORITY
from terminal_requests import run_onboard, run_slack_add
from aws_users import run_aws_user


def _add_jira_shared_args(subparser: argparse.ArgumentParser) -> None:
    subparser.add_argument(
        "--user",
        metavar="\"FIRST LAST EMAIL[, ...]\"",
        help="Comma-separated users (email must end with @global.komatsu)",
    )
    subparser.add_argument("--users-file", metavar="CSV", help="CSV with columns: first_name, last_name, email")
    subparser.add_argument(
        "--location", default=DEFAULT_LOCATION, choices=list(LOCATIONS.keys()),
        help=f"Office location (default: {DEFAULT_LOCATION})",
    )
    subparser.add_argument(
        "--priority", default=DEFAULT_PRIORITY, choices=list(PRIORITIES.keys()),
        help=f"Request priority (default: {DEFAULT_PRIORITY})",
    )
    subparser.add_argument("--summary", default=None, help="Override the default summary")
    subparser.add_argument("--description", default=None, help="Override the default description")
    subparser.add_argument("--participants", metavar="ACCOUNT_ID", nargs="+")


def main():
    parser = argparse.ArgumentParser(
        description="Komatsu onboarding automation.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=f"Valid locations: {', '.join(LOCATIONS.keys())}",
    )
    parser.add_argument("--requestor-email", default=os.environ.get("ATLASSIAN_EMAIL", ""))
    parser.add_argument("--api-token", default=os.environ.get("KATANA_ONBOARDER_API", ""))

    subparsers = parser.add_subparsers(dest="command", required=True)

    # -- onboard --------------------------------------------------------------
    onboard_parser = subparsers.add_parser(
        "onboard",
        help="Provision Okta + Slack for new Komatsu engineers",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Example:
  python3 create_service_desk_request.py onboard \\
    --user "Taro Yamamoto taro.yamamoto@global.komatsu"
""",
    )
    _add_jira_shared_args(onboard_parser)
    onboard_parser.add_argument(
        "--channels",
        metavar="\"#channel1[, ...]\"",
        default=None,
        help="Additional Slack channels to include beyond the defaults",
    )

    # -- slack-add ------------------------------------------------------------
    slack_parser = subparsers.add_parser(
        "slack-add",
        help="Add Komatsu engineers to additional Slack channels",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Example:
  python3 create_service_desk_request.py slack-add \\
    --user "Taro Yamamoto taro.yamamoto@global.komatsu" \\
    --channels "#ext-program-katana, #general"
""",
    )
    _add_jira_shared_args(slack_parser)
    slack_parser.add_argument(
        "--channels", required=True,
        metavar="\"#channel1[, ...]\"",
        help="Comma-separated Slack channels to add the users to",
    )

    # -- aws-user -------------------------------------------------------------
    aws_parser = subparsers.add_parser(
        "aws-user",
        help="Create IAM user (first.last), add to katana group, email credentials",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Example:
  python3 create_service_desk_request.py aws-user \\
    --user "Taro Yamamoto taro.yamamoto@global.komatsu"

  # Send credentials to a different address
  python3 create_service_desk_request.py aws-user \\
    --user "Taro Yamamoto taro.yamamoto@global.komatsu" \\
    --send-to manager@applied.co
""",
    )
    aws_parser.add_argument(
        "--user",
        metavar="\"FIRST LAST EMAIL[, ...]\"",
        help="Comma-separated users. Email is where credentials are sent by default.",
    )
    aws_parser.add_argument("--users-file", metavar="CSV", help="CSV with columns: first_name, last_name, email")
    aws_parser.add_argument(
        "--send-to", metavar="EMAIL", default=None,
        help="Override destination for all credential emails",
    )

    args = parser.parse_args()

    if args.command in ("onboard", "slack-add"):
        if not args.requestor_email or not args.api_token:
            parser.error("Provide --requestor-email and --api-token, or set ATLASSIAN_EMAIL and KATANA_ONBOARDER_API.")
        if not args.user and not args.users_file:
            parser.error("Provide at least one --user or a --users-file.")
        if args.command == "onboard":
            run_onboard(args, parser)
        else:
            run_slack_add(args, parser)

    elif args.command == "aws-user":
        if not args.user and not args.users_file:
            parser.error("Provide at least one --user or a --users-file.")
        run_aws_user(args, parser)


if __name__ == "__main__":
    main()
