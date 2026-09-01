# Interact test server provisioning

Ansible setup for the Interact (Wayfarer) test box at `49.12.121.62`
(`interact-test.bcc.media`, Debian 13). **Warning:** this box doubles as the
loadtest bench (see `notes` about `make loadtest-remote-*`); running this
playbook reconfigures Postgres and the firewall on it.

## What it does

| Role        | Purpose |
| ----------- | ------- |
| `hardening` | apt upgrade + unattended security upgrades, sshd hardening (key-only, root stays allowed with keys for existing tooling), fail2ban, sysctl hardening |
| `tuning`    | load-test base tuning: nofile 65535 (limits.d + systemd default), somaxconn/syn-backlog 8192, wider ephemeral port range, tcp_tw_reuse, nf_conntrack_max 131072 (ufw stateful tracking), CPU governor pinned to `performance` |
| `firewall`  | ufw: deny incoming by default; allow 22 (rate-limited), 80 (ACME http-01), 443 (native TLS) |
| `postgres`  | PostgreSQL 17 from Debian repos, scram-only auth, loopback-only, tuning derived from host RAM/CPUs, pg_stat_statements |
| `interact`  | Creates the `interact` database and role for the app |
| `wayfarer`  | Native (proxyless) blue/green deploy layout: `wayfarer@{blue,green}` systemd units sharing the port via SO_REUSEPORT, per-color admin/health ports (9441/9442), split DB pools, `bin/deploy.sh` invoked by Semaphore CI (`.semaphore/`). Secret env `/opt/wayfarer/wayfarer.env` is placed manually |

## Prerequisites

```sh
ansible-galaxy collection install -r requirements.yml
```

DNS for `interact-test.bcc.media` must point at `49.12.121.62` before the
wayfarer server first starts with TLS enabled, or ACME issuance will fail.

## Secrets

The Interact DB password must be provided; the playbook refuses the
`CHANGE_ME` placeholder. Either:

```sh
# one-off
ansible-playbook site.yml -e interact_db_password=...

# or with vault: put vault_interact_db_password in an encrypted vars file
ansible-vault create group_vars/interact/vault.yml
ansible-playbook site.yml --ask-vault-pass
```

## Run

```sh
make deps      # once: install ansible collections
make ping      # connectivity check
make dry-run   # --check --diff
make deploy    # everything
make postgres  # any single role by name (hardening/tuning/firewall/postgres/interact/wayfarer)
```

The Makefile picks up the DB password from `$INTERACT_DB_PASSWORD` if set;
otherwise use vault as described above.

## Design notes / gotchas

- **TLS lives in the app:** the wayfarer server terminates TLS natively on
  443 and answers ACME http-01 challenges on 80 (certs self-renew under
  `/opt/wayfarer/autocert`). No proxy in front.
- **App → Postgres:** the app runs natively on the box and reaches the
  database at `127.0.0.1:5432`, database/user `interact`, scram auth.
  Postgres listens on loopback only.
- **Dokploy/Docker removal:** this playbook no longer installs Dokploy, but
  it doesn't uninstall an existing install either. On a box previously
  provisioned with it, tear down Dokploy/Traefik/Docker swarm by hand (or
  re-image) — otherwise Traefik still holds 80/443 and blocks the app.
  Watch for host-local leftovers that assumed the Docker bridge, e.g. a
  `loadtest.env` pointing at `host.docker.internal:5432` (now `127.0.0.1`).
- **Postgres tuning** is computed from Ansible facts at run time
  (25% RAM shared_buffers, SSD planner costs, parallelism per vCPU) in
  `roles/postgres/templates/90-tuning.conf.j2`. Changing `max_connections`
  requires a Postgres restart (the role's handler does this).
- The existing bench setup (`wayfarer.service`, `bench` DB user) is left
  untouched by this playbook, but the pg_hba rewrite is opinionated — if the
  bench DB auth breaks, add its entries to
  `roles/postgres/templates/pg_hba.conf.j2`.
