# Maintenance Cron Jobs

What to schedule, and where. Both need `Authorization: Bearer <API key>`.

| Endpoint | What it does |
|---|---|
| `POST /api/maintenance/sync-user-data` | Refreshes existing users' data from Members. No `?limit=` = whole table. |
| `POST /api/maintenance/import-new-members` | Creates users for newly-eligible members (e.g. just turned 12). |

## Where they are scheduled

On the prod box, via `ansible/roles/jobs` (systemd timers, schedules in
`wayfarer_jobs` in `ansible/group_vars/all.yml`, all Europe/Oslo):

| Job | Schedule | Calls |
|---|---|---|
| `export-translations` | hourly at :00 | `POST /api/translations/export/all` (`X-Export-Key: $TRANSLATIONS_EXPORT_KEY`) |
| `sync-ssf` | daily 05:00 | `POST /ssf/sync/hidden-treasures-podcast` (`X-Sync-Key: $SSF_SYNC_KEY`) |
| `sync-members` | Tue 02:00 (Monday night) | `sync-user-data`, then `import-new-members` (Bearer key from the `cron:` entry in `EXTERNAL_API_KEYS`) |

All keys are read from `/opt/wayfarer/wayfarer.env` at run time. Logs:
`journalctl -u wayfarer-job-<name>`. These replaced the Google Cloud
Scheduler jobs; the trigger endpoints only accept POST (a GET is a 404).
