# Maintenance Cron Jobs

What to schedule, and where. Both need `Authorization: Bearer <API key>`.

| Endpoint | What it does |
|---|---|
| `POST /api/maintenance/sync-user-data` | Refreshes existing users' data from Members. No `?limit=` = whole table. |
| `POST /api/maintenance/import-new-members` | Creates users for newly-eligible members (e.g. just turned 12). |
