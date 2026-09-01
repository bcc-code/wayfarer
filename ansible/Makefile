# Provisioning for the Interact prod server (49.12.121.62).
# Requires: ansible-core (brew install ansible), plus `make deps` once.
#
# The interact role needs a DB password, and the tailscale role needs an
# auth key for the first tailnet join. Either export them:
#   export INTERACT_DB_PASSWORD=...
#   export TAILSCALE_AUTHKEY=tskey-auth-...   # first join only
# or use an ansible-vault file (see README.md).

PLAYBOOK := site.yml
EXTRA_VARS := $(if $(INTERACT_DB_PASSWORD),-e interact_db_password=$(INTERACT_DB_PASSWORD),)
EXTRA_VARS += $(if $(TAILSCALE_AUTHKEY),-e tailscale_authkey=$(TAILSCALE_AUTHKEY),)

.PHONY: help deps check lint dry-run deploy ping hardening tuning firewall tailscale postgres interact wayfarer

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

deps: ## Install required Ansible collections
	ansible-galaxy collection install -r requirements.yml

ping: ## Check connectivity to the box
	ansible all -m ping

check: ## Syntax-check the playbook
	ansible-playbook $(PLAYBOOK) --syntax-check

lint: ## Run ansible-lint (if installed)
	ansible-lint $(PLAYBOOK)

dry-run: check ## Full dry run with diff output
	ansible-playbook $(PLAYBOOK) $(EXTRA_VARS) --check --diff

deploy: check ## Provision the server (everything)
	ansible-playbook $(PLAYBOOK) $(EXTRA_VARS)

# Per-role targets, e.g. `make postgres`
hardening tuning firewall tailscale postgres interact wayfarer: check ## Run a single role by tag
	ansible-playbook $(PLAYBOOK) $(EXTRA_VARS) -t $@
