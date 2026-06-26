"""
Jira Service Management requests for Komatsu onboarding.
Handles onboard (type 244) and slack-add (type 307) flows.
"""

import csv
import sys
import argparse
import requests
from base64 import b64encode
from dataclasses import dataclass
from typing import Any

from constants import (
    BASE_URL,
    SERVICE_DESK_ID,
    REQUEST_TYPE_ONBOARD,
    REQUEST_TYPE_SLACK,
    LOCATIONS,
    PRIORITIES,
    ONBOARD_DEFAULT_SUMMARY,
    ONBOARD_DESCRIPTION_TEMPLATE,
    SLACK_DEFAULT_SUMMARY,
    SLACK_DESCRIPTION_TEMPLATE,
)

KOMATSU_EMAIL_DOMAIN = "@global.komatsu"


# ---------------------------------------------------------------------------
# Shared data model (also used by aws_users.py)
# ---------------------------------------------------------------------------

@dataclass
class Engineer:
    first_name: str
    last_name: str
    email: str

    @property
    def aws_username(self) -> str:
        return f"{self.first_name.lower()}.{self.last_name.lower()}"


# ---------------------------------------------------------------------------
# Input parsing
# ---------------------------------------------------------------------------

def validate_komatsu_email(email: str, source: str, parser: argparse.ArgumentParser) -> None:
    if not email.endswith(KOMATSU_EMAIL_DOMAIN):
        parser.error(f"Invalid email {email!r} ({source}): must end with {KOMATSU_EMAIL_DOMAIN}")


def parse_users(value: str, parser: argparse.ArgumentParser) -> list[Engineer]:
    engineers = []
    for entry in value.split(","):
        parts = entry.strip().split()
        if len(parts) != 3:
            parser.error(f"Each user entry must be 'First Last email', got: {entry.strip()!r}")
        first, last, email = parts
        validate_komatsu_email(email, "--user", parser)
        engineers.append(Engineer(first_name=first, last_name=last, email=email))
    return engineers


def load_users_file(path: str, parser: argparse.ArgumentParser) -> list[Engineer]:
    engineers = []
    with open(path, newline="") as f:
        sample = f.read(1024)
        f.seek(0)
        has_header = csv.Sniffer().has_header(sample)
        if has_header:
            reader = csv.DictReader(f)
            for row in reader:
                email = row["email"].strip()
                validate_komatsu_email(email, path, parser)
                engineers.append(Engineer(
                    first_name=row["first_name"].strip(),
                    last_name=row["last_name"].strip(),
                    email=email,
                ))
        else:
            for i, row in enumerate(csv.reader(f), start=1):
                if len(row) != 3:
                    parser.error(f"{path} row {i}: expected 3 columns (first_name, last_name, email), got {len(row)}")
                first, last, email = (c.strip() for c in row)
                validate_komatsu_email(email, path, parser)
                engineers.append(Engineer(first_name=first, last_name=last, email=email))
    return engineers


# ---------------------------------------------------------------------------
# Jira API
# ---------------------------------------------------------------------------

def _headers(email: str, api_token: str) -> dict:
    token = b64encode(f"{email}:{api_token}".encode()).decode()
    return {
        "Authorization": f"Basic {token}",
        "Content-Type": "application/json",
        "Accept": "application/json",
        "X-ExperimentalApi": "opt-in",
    }


def onboard_description(engineers: list[Engineer], extra_channels: list[str] | None = None) -> str:
    user_lines = "\n".join(f"{e.first_name} {e.last_name} | {e.email}" for e in engineers)
    desc = ONBOARD_DESCRIPTION_TEMPLATE.format(user_lines=user_lines)
    if extra_channels:
        desc += "\n" + "\n".join(extra_channels)
    return desc


def slack_description(engineers: list[Engineer], channels: list[str]) -> str:
    user_lines = "\n".join(f"{e.first_name} {e.last_name} | {e.email}" for e in engineers)
    channel_lines = "\n".join(channels)
    return SLACK_DESCRIPTION_TEMPLATE.format(user_lines=user_lines, channel_lines=channel_lines)


def post_request(
    auth_email: str,
    api_token: str,
    request_type_id: str,
    summary: str,
    description: str,
    location: str,
    priority: str,
    participants: list[str] | None = None,
) -> dict:
    payload: dict[str, Any] = {
        "serviceDeskId": SERVICE_DESK_ID,
        "requestTypeId": request_type_id,
        "requestFieldValues": {
            "summary": summary,
            "description": description,
            "customfield_10513": {"id": LOCATIONS[location]},
            "priority": {"id": PRIORITIES[priority]},
        },
    }
    if participants:
        payload["requestParticipants"] = participants

    resp = requests.post(
        BASE_URL + "/rest/servicedeskapi/request",
        headers=_headers(auth_email, api_token),
        json=payload,
        timeout=30,
    )
    if not resp.ok:
        print(f"ERROR {resp.status_code}: {resp.text}", file=sys.stderr)
        resp.raise_for_status()
    return resp.json()


def resolve_link(result: dict) -> str:
    return result.get("_links", {}).get("web") or (
        f"{BASE_URL}/browse/{result['issueKey']}" if result.get("issueKey") else ""
    )


def run_onboard(args: argparse.Namespace, parser: argparse.ArgumentParser) -> None:
    engineers = _collect_engineers(args, parser)
    extra_channels = [c.strip() for c in args.channels.split(",")] if args.channels else None
    summary = args.summary or ONBOARD_DEFAULT_SUMMARY
    description = args.description or onboard_description(engineers, extra_channels)
    result = post_request(
        auth_email=args.requestor_email,
        api_token=args.api_token,
        request_type_id=REQUEST_TYPE_ONBOARD,
        summary=summary,
        description=description,
        location=args.location,
        priority=args.priority,
        participants=args.participants,
    )
    link = resolve_link(result)
    print(link if link else result)


def run_slack_add(args: argparse.Namespace, parser: argparse.ArgumentParser) -> None:
    engineers = _collect_engineers(args, parser)
    channels = [c.strip() for c in args.channels.split(",")]
    summary = args.summary or SLACK_DEFAULT_SUMMARY
    description = args.description or slack_description(engineers, channels)
    result = post_request(
        auth_email=args.requestor_email,
        api_token=args.api_token,
        request_type_id=REQUEST_TYPE_SLACK,
        summary=summary,
        description=description,
        location=args.location,
        priority=args.priority,
        participants=args.participants,
    )
    link = resolve_link(result)
    print(link if link else result)


def _collect_engineers(args: argparse.Namespace, parser: argparse.ArgumentParser) -> list[Engineer]:
    engineers: list[Engineer] = []
    if args.user:
        engineers += parse_users(args.user, parser)
    if args.users_file:
        engineers += load_users_file(args.users_file, parser)
    return engineers
