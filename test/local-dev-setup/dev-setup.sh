#!/usr/bin/env bash
# AIM Local Development Test Environment - Bash/Zsh
# Usage: source test/local-dev-setup/dev-setup.sh [command]
#
# Examples:
#   source test/local-dev-setup/dev-setup.sh              # Initialize and load
#   source test/local-dev-setup/dev-setup.sh update       # Update and load

eval $(./test/local-dev-setup/local-dev.sh "$@")
