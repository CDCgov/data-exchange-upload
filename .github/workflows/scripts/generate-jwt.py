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

print(os.environ["PRIVATE_KEY"].encode('ascii'))

# Create JWT
encoded_jwt = jwt.encode(
    payload, os.environ["PRIVATE_KEY"].encode('ascii'), algorithm='RS256')

print(encoded_jwt)
