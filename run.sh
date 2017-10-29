#!/bin/bash

_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
_PID="${_PATH}/run.pid"
_LOG="${_PATH}/run.log"
_PROGRAM="${_PATH}/8xio"

cd "${_PATH}"

if [ -e "${_PID}" ] && (ps -u $(whoami) -opid= | grep -P "^\s*$(cat ${_PID})$" &> /dev/null); then
  echo "Program '${_PROGRAM}' is already running."
  exit 99
fi

"${_PROGRAM}" >> "${_LOG}" &

echo $! > "${_PID}"
chmod 644 "${_PID}"
