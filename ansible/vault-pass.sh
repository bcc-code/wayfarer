#!/bin/sh
# Ansible vault password client: reads the password from 1Password
# ("Interact Ansible Vault") with the op CLI, so playbook runs don't prompt.
# Wired up via vault_password_file in ansible.cfg. Override the item with
# OP_ANSIBLE_VAULT_ITEM (ID or name) if needed.
set -eu
exec op item get "${OP_ANSIBLE_VAULT_ITEM:-2mbbsu5zrote24aszubwu6ifrq}" --fields label=password --reveal
