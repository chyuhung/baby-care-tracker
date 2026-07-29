# Baby Care Tracker

## Goal
Family group sharing, timezone fix, record performance optimization

## Progress
### Done
- Family features: routes, checkBabyFamily helper, family API on frontend, family management UI in ProfilePage
- Login fix: username input hidden by `v-if="isRegister"` — always show in both modes
- Timezone: client sends `X-Timezone-Offset` header; backend uses Go-calculated UTC ranges instead of SQLite `localtime` modifier; `nowDatetime()` returns local time, `utcToLocalDatetime()` for edit loading
- Backend record sorting: `parseTime()` helper converts strings to `time.Time` for correct sort (handles RFC3339 and legacy local-time strings)
- Record performance: WebSocket dispatches `record-created` / `record-deleted` events with payload; pages update local array directly instead of re-fetching all records
- Trend chart: migrated to Go-side aggregation using `parseTime()` + `time.FixedZone` — fetches raw records, groups by local date, avoids SQLite `date()` timezone quirks
- Time display: replaced relative times ("X小时前") with absolute format (`MM-DD HH:mm`) in home/timeline; removed `tick` timer
- Sleep + temperature features: 2 new tables (`sleep_records`, `temperature_records`); sleep start/stop with ongoing tracking; temperature measurement with 5 locations & fever threshold
- Dedicated pages: SleepPage (timer, today summary, edit started_at/ended_at), TemperaturePage (quick record, latest reading, location selector), TrendPage (4-chart aggregation)
- HomePage 4-card grid: 2×2 layout with feeding/diaper/sleep/temperature cards; inline sleep start/stop; temperature fever color display
- Naming conventions aligned: `记录睡眠` / `记录体温` page titles, `记录` submit button, `更新记录` edit button, icons-only card entry buttons

### Known Issues
- `vue-tsc` typecheck fails on Node.js v24 — not a code issue
- No `.gitignore` exists; `baby-care-tracker.exe` binary is tracked in git

## Architecture
- **Backend**: Go + gin on `:8080`, modernc.org/sqlite, WebSocket broadcast via Hub pattern
- **Frontend**: Vue 3 + Pinia + Vue Router, Vite dev server on `:5173`
- **DB**: `C:\app\data\app.db`; DATETIME columns always return UTC with `Z` suffix

## Key Decisions
- Timezone: client offset header, not server TZ; stored times are UTC
- Record events: `record-created` and `record-deleted` CustomEvents on `window` — WebSocket broadcasts trigger these; delete handlers dispatch directly for instant UI
