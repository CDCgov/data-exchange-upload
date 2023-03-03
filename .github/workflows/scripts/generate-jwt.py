import argparse
import jwt
import time
import os

parser = argparse.ArgumentParser()
parser.add_argument('--app-id', required=True)

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

print(f'jwt={encoded_jwt}')
