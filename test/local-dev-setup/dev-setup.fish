#!/usr/bin/env fish
# AIM Local Development Test Environment - Fish Shell
# Usage: source test/local-dev-setup/dev-setup.fish [command]
#
# Examples:
#   source test/local-dev-setup/dev-setup.fish              # Initialize and load
#   source test/local-dev-setup/dev-setup.fish update       # Update and load

eval (./test/local-dev-setup/local-dev.sh $argv)
