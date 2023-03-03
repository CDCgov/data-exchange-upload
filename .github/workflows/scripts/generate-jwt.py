import argparse
import jwt
import time

parser = argparse.ArgumentParser()
parser.add_argument('--private-key', required=True)
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
encoded_jwt = jwt.encode(payload, args.private_key, algorithm='RS256')

print(f"JWT:  ", encoded_jwt)
