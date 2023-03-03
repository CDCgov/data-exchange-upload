import argparse
import jwt
import time

parser = argparse.ArgumentParser()
parser.add_argument('--pem', required=True)
parser.add_argument('--app-id', required=True)

args = parser.parse_args()

# Open PEM
with open(args.pem, 'rb') as pem_file:
    signing_key = jwt.jwk_from_pem(pem_file.read())

payload = {
    # Issued at time
    'iat': int(time.time()),
    # JWT expiration time (10 minutes maximum)
    'exp': int(time.time()) + 600,
    # GitHub App's identifier
    'iss': args.app_id
}

# Create JWT
encoded_jwt = jwt.encode(payload, signing_key, algorithm='RS256')

print(encoded_jwt)
