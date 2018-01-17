#!/bin/bash

_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
_PID="${_PATH}/run.pid"
_LOG="${_PATH}/run.log"
_PROGRAM="${_PATH}/8xio"

cd "${_PATH}"
go build
kill -9 `cat ${_PID}`
./run.sh
tail -f ${_LOG}
