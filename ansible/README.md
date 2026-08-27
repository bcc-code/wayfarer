# Interact test server provisioning

Ansible setup for the Interact (Wayfarer) test box at `49.12.121.62`
(`interact-test.bcc.media`, Debian 13). **Warning:** this box doubles as the
loadtest bench (see `notes` about `make loadtest-remote-*`); running this
playbook reconfigures Postgres and the firewall on it.

## What it does

| Role        | Purpose |
| ----------- | ------- |
| `hardening` | apt upgrade + unattended security upgrades, sshd hardening (key-only, root stays allowed with keys for existing tooling), fail2ban, sysctl hardening |
| `tuning`    | load-test base tuning: nofile 65535 (limits.d + systemd default), somaxconn/syn-backlog 8192, wider ephemeral port range, tcp_tw_reuse, nf_conntrack_max 262144 for Docker NAT |
| `firewall`  | ufw: deny incoming by default; allow 22 (rate-limited), 80, 443; Dokploy UI on 3000 (optionally source-restricted); Postgres 5432 only from Docker subnets |
| `postgres`  | PostgreSQL 17 from Debian repos, scram-only auth, listens on loopback + Docker bridge, tuning derived from host RAM/CPUs, pg_stat_statements |
| `interact`  | Creates the `interact` database and role for the app |
| `dokploy`   | Installs Dokploy via the official installer; its bundled Traefik owns 80/443 and terminates TLS |

## Prerequisites

```sh
ansible-galaxy collection install -r requirements.yml
```

DNS for `interact-test.bcc.media` must point at `49.12.121.62` before
assigning the domain to the app in Dokploy, or ACME issuance will fail.

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
make dokploy   # any single role by name (hardening/firewall/postgres/interact/dokploy)
```

The Makefile picks up the DB password from `$INTERACT_DB_PASSWORD` if set;
otherwise use vault as described above.

## Design notes / gotchas

- **Domains/TLS live in Dokploy:** its bundled Traefik owns 80/443. After
  provisioning, open the Dokploy UI (port 3000), create the Interact app,
  and assign the domain `interact-test.bcc.media` with HTTPS (Let's
  Encrypt) enabled — Traefik handles certificates and routing.
- **Docker vs ufw:** Docker publishes container ports via iptables NAT,
  *bypassing* ufw. Don't publish extra app ports in Dokploy — let Traefik
  route to the container over the Docker network.
- **App → Postgres:** the app container reaches the host database at
  `172.17.0.1:5432` (Docker bridge gateway), database/user `interact`,
  scram auth. pg_hba only admits that user/database pair from Docker subnets.
- **Postgres tuning** is computed from Ansible facts at run time
  (25% RAM shared_buffers, SSD planner costs, parallelism per vCPU) in
  `roles/postgres/templates/90-tuning.conf.j2`.
- The existing bench setup (`wayfarer.service`, `bench` DB user) is left
  untouched by this playbook, but the pg_hba rewrite is opinionated — if the
  bench DB auth breaks, add its entries to
  `roles/postgres/templates/pg_hba.conf.j2`.
