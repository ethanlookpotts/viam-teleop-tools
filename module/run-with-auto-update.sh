#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo 'Usage: ./run-with-auto-update.sh 

    Special script to get auto-updating behavior for local modules from pushes to main.

    This runs `git pull` then builds the module then starts the module.
    Then, this script kills itself every 5 minutes and is expected to be restarted by the rover.
'
    exit
fi

script_dir="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"

pushd "$script_dir/.."
git pull
just build
timeout 300 ./dist/module "$1"
