#!/bin/bash

##############################################################################################
# Based on https://gist.github.com/eisenreich/195ab1f05715ec86e300f75d007d711c
#
# Wait until URL returns HTTP 200
#
# Example: ./wait_for_it.sh  "http://192.168.56.101:8080"
##############################################################################################

wait-for-url() {
  timeout --foreground -s TERM 60s bash -c \
      'while [[ "$(curl -s -o /dev/null -m 3 -L -w ''%{http_code}'' ${0})" != "200" ]];\
      do sleep 2;\
      done' ${1}
  local TIMEOUT_RETURN="$?"
  if [[ "${TIMEOUT_RETURN}" == 0 ]]; then
      return 
  elif [[ "${TIMEOUT_RETURN}" == 124 ]]; then
      exit "${TIMEOUT_RETURN}"
  else
      exit "${TIMEOUT_RETURN}"
  fi
}

wait-for-url "$@"
