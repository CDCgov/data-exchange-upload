#!/bin/bash

##############################################################################################
# Based on https://gist.github.com/eisenreich/195ab1f05715ec86e300f75d007d711c
#
# Wait until URL returns HTTP 200
#
# Example: ./wait_for_it.sh "http://192.168.56.101:8080"
##############################################################################################

wait-for-url() {
    echo "Testing: $1"
    timeout --foreground -s TERM 60s bash -c \
        'while [[ "$(curl -s -o /dev/null -m 3 -L -w ''%{http_code}'' ${0})" != "200" ]];\
        do echo "Waiting for ${0}" && sleep 2;\
        done' ${1}
    local TIMEOUT_RETURN="$?"
    if [[ "${TIMEOUT_RETURN}" == 0 ]]; then
        echo "OK: ${1}"
        return 
    elif [[ "${TIMEOUT_RETURN}" == 124 ]]; then
        echo "TIMEOUT: ${1} -> EXIT"
        exit "${TIMEOUT_RETURN}"
    else
        echo "Other error with code ${TIMEOUT_RETURN}: ${1} -> EXIT"
        exit "${TIMEOUT_RETURN}"
    fi
}

echo "Wait for URL: $@"
wait-for-url "$@"
echo ""
echo "SUCCESSFUL"
