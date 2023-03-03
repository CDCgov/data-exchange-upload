#!/bin/bash

# Check that both arguments are present
if [ $# -ne 2 ]; then
    echo "Usage: $0 <jwt> <installation_id>"
    exit 1
fi

# Assign the arguments to variables
jwt=$1
installation_id=$2

token=$(curl -i -X POST \
-H "Authorization: Bearer ${jwt}" \
-H "Accept: application/vnd.github+json" \
"https://api.github.com/app/installation/${installation_id}/access_tokens" | jq '.token')

echo installToken=${token}
