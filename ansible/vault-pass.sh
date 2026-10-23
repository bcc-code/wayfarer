#!/bin/sh
# Ansible vault password client: reads the password from 1Password
# ("Interact Ansible Vault") with the op CLI, so playbook runs don't prompt.
# Wired up via vault_password_file in ansible.cfg. Overrides:
#   OP_ANSIBLE_VAULT_ITEM     item ID or name
#   OP_ANSIBLE_VAULT_VAULT    1Password vault holding the item
#   OP_ANSIBLE_VAULT_ACCOUNT  1Password account (sign-in address or ID), when
#                             several accounts are signed in
set -eu
item="${OP_ANSIBLE_VAULT_ITEM:-2mbbsu5zrote24aszubwu6ifrq}"
vault="${OP_ANSIBLE_VAULT_VAULT:-}"
account="${OP_ANSIBLE_VAULT_ACCOUNT:-bcc-no.1password.eu}"

set -- item get "$item" --fields label=password --reveal
[ -n "$vault" ] && set -- "$@" --vault "$vault"
[ -n "$account" ] && set -- "$@" --account "$account"
exec op "$@"
