"""
AWS IAM user creation for Komatsu onboarding.
Creates first.last IAM users, adds to katana group, emails temporary credentials via SES.
"""

import csv
import os
import secrets
import string
import sys
import argparse

import boto3
from botocore.exceptions import ClientError

from constants import AWS_ACCOUNT_ID, AWS_IAM_GROUP, AWS_REGION, AWS_CONSOLE_URL
from terminal_requests import Engineer


# ---------------------------------------------------------------------------
# Input parsing (no Komatsu email validation - email is credential recipient)
# ---------------------------------------------------------------------------

def parse_aws_users(value: str, parser: argparse.ArgumentParser) -> list[Engineer]:
    engineers = []
    for entry in value.split(","):
        parts = entry.strip().split()
        if len(parts) != 3:
            parser.error(f"Each user entry must be 'First Last email', got: {entry.strip()!r}")
        first, last, email = parts
        engineers.append(Engineer(first_name=first, last_name=last, email=email))
    return engineers


def load_aws_users_file(path: str) -> list[Engineer]:
    engineers = []
    with open(path, newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            engineers.append(Engineer(
                first_name=row["first_name"].strip(),
                last_name=row["last_name"].strip(),
                email=row["email"].strip(),
            ))
    return engineers


# ---------------------------------------------------------------------------
# IAM user creation
# ---------------------------------------------------------------------------

def _generate_temp_password(length: int = 16) -> str:
    specials = "!@#$%^&*()"
    alphabet = string.ascii_letters + string.digits + specials
    while True:
        pw = "".join(secrets.choice(alphabet) for _ in range(length))
        if (any(c.islower() for c in pw)
                and any(c.isupper() for c in pw)
                and any(c.isdigit() for c in pw)
                and any(c in specials for c in pw)):
            return pw


def create_iam_user(engineer: Engineer) -> dict:
    """
    Create an IAM user (first.last), add to katana group,
    set a temporary password, and generate access keys.
    Returns a dict of credentials.
    """
    iam = boto3.client("iam")
    username = engineer.aws_username

    try:
        iam.create_user(UserName=username)
        print(f"Created IAM user: {username}")
    except ClientError as e:
        if e.response["Error"]["Code"] == "EntityAlreadyExists":
            print(f"WARNING: IAM user '{username}' already exists, skipping creation.", file=sys.stderr)
        else:
            raise

    iam.add_user_to_group(GroupName=AWS_IAM_GROUP, UserName=username)
    print(f"Added '{username}' to group '{AWS_IAM_GROUP}'.")

    temp_password = _generate_temp_password()
    try:
        iam.create_login_profile(
            UserName=username,
            Password=temp_password,
            PasswordResetRequired=True,
        )
    except ClientError as e:
        if e.response["Error"]["Code"] == "EntityAlreadyExists":
            iam.update_login_profile(
                UserName=username,
                Password=temp_password,
                PasswordResetRequired=True,
            )
        else:
            raise

    return {
        "username": username,
        "temp_password": temp_password,
    }


# ---------------------------------------------------------------------------
# Email delivery
# ---------------------------------------------------------------------------

def send_credentials_email(engineer: Engineer, credentials: dict, send_to: str) -> None:
    """Send AWS credentials to send_to via SES."""
    sender = os.environ.get("SES_SENDER_EMAIL", "noreply@applied.co")
    ses = boto3.client("ses", region_name=AWS_REGION)

    subject = f"AWS Account Credentials - {credentials['username']}"
    body = f"""\
Hello {engineer.first_name},

Your AWS account for the Katana project has been created. Your credentials are below.

Console Login URL : {AWS_CONSOLE_URL}
Username          : {credentials['username']}
Temporary Password: {credentials['temp_password']}

You will be required to change your password on first login.

For more information on AWS developer environment setup, see here:
https://docs.google.com/document/d/1E8tP3sscUJWwjokBH3xLBTBup_B7YPmFh-XovTFGKPE/edit?tab=t.0#heading=h.g9kv7nc26b63
Note: you need to have your Google account set up in order to view.

Please store these credentials securely and do not share them.

Best regards,
Katana Onboarding Team
"""

    ses.send_email(
        Source=sender,
        Destination={"ToAddresses": [send_to]},
        Message={
            "Subject": {"Data": subject},
            "Body": {"Text": {"Data": body}},
        },
    )
    print(f"Credentials emailed to {send_to}.")


# ---------------------------------------------------------------------------
# Command runner
# ---------------------------------------------------------------------------

def run_aws_user(args: argparse.Namespace, parser: argparse.ArgumentParser) -> None:
    engineers: list[Engineer] = []
    if args.user:
        engineers += parse_aws_users(args.user, parser)
    if args.users_file:
        engineers += load_aws_users_file(args.users_file)

    for engineer in engineers:
        credentials = create_iam_user(engineer)
        recipient = args.send_to or engineer.email
        send_credentials_email(engineer, credentials, send_to=recipient)
        print(f"Done: {engineer.aws_username}")
