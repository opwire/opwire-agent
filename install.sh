#!/usr/bin/env bash

set -e

usage() {
  echo "Usage: curl https://opwire.org/opwire-agent/install.sh | sudo bash" 1>&2;
  exit 1;
}

#check for help flag
if [ -n "$1" ] && [ "$1" == "help" ]; then
    usage
fi

