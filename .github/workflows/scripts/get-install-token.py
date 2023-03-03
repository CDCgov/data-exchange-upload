import argparse
import jwt
import time
import os
import requests

parser = argparse.ArgumentParser()
parser.add_argument('--app-id', required=True)
parser.add_argument('--installation-id', required=True)

args = parser.parse_args()

payload = {
    # Issued at time
    'iat': int(time.time()),
    # JWT expiration time (10 minutes maximum)
    'exp': int(time.time()) + 600,
    # GitHub App's identifier
    'iss': args.app_id
}

# Create JWT
encoded_jwt = jwt.encode(payload, os.environ["PRIVATE_KEY"], algorithm='RS256')

# Get installation auth token
url = f'https://api.github.com/app/installations/{args.installation_id}/access_tokens'
headers = {
    "Authorization": f'Bearer {encoded_jwt}',
    "Accept": "application/vnd.github+json"
}

token = requests.post(url, headers=headers).json()['token']

print(f'token={token}')
