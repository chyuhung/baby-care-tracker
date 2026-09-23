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
- Supplement feature: `supplement_records` table (name, dosage_value, dosage_unit, note); full record lifecycle (create/update/delete/timeline filter/RecordCard); HomePage full-width card (今日次数/距上次/平均间隔 + add); TrendPage `supplement` category (count bar/scatter + summary cards); theme color `--supplement` violet; PWA cache bumped to v10
- iOS three-zone chrome: sticky frosted-glass headers + 0.5px hairline on tab bar/FormBar (`hairline-top`/`hairline-bottom`); page transition simplified to opacity-fade-only (transform layer eats first tap)
- iOS tap/gesture fixes: removed `user-scalable=no`/`maximum-scale` from viewport (PWA standalone keyboard focus) and global `overscroll-behavior: none` (double-tap needed); then replaced PullRefresh with Pointer Events state machine (UIRefreshControl-equivalent): single pointerId + capture, live `container.scrollTop` check (no snapshot/`lockedUp` lock), 10px tap slop, `overflow-y-auto + overscroll-contain + touch-pan-y` on container, iOS-style indicator (content stays put, no whole-page transform)
- Single scroll container architecture: 6 PullRefresh pages (Home/Timeline/Trend/Profile/Supplement/Temperature) now `h-dvh` flex col with PullRefresh as the inner scroller; page headers moved into PullRefresh `#header` slot (sticky inside scroller, glass blur passes under)
- Growth standard WS/T 423-2022: embedded percentile tables (`backend/data/wst423_2022.json` + `growthstd.go`, weight/height 0–81月, head 0–36月, male/female, P3–P97); `growthPercentile` percentile lookup with segment-linear interpolation + endpoint extrapolation (replaced WHO LMS/z-scores); new `GET /babies/:id/growth/reference` endpoint returns P3/P25/P50/P75/P97 reference curves
- GrowthPage chart: age-in-months X axis (not dates), hospital-style red/yellow/green reference zones (green P25–P75, yellow P3–P25/P75–P97, red <P3/>P97) + legend + WS/T 423-2022 footnote; `pctClass` thresholds aligned to the standard's 5-grade evaluation (<3/>97 danger, <25/>75 warning)
- LargeTitleNav iOS behavior: two-root fragment — sticky compact bar (h-11 glass, inline title fades in past `COLLAPSE_AT=30`/out below `EXPAND_AT=8`, `#actions` + `#filters` always pinned) + flowing large-title block (`<h1>` 34px + `#sub`, in normal flow, scrolls under the opaque bar); `large=false` renders compact-only; PullRefresh `measureHeader` now sums header-slot children before content ref (supports multi-root fragments, replaces `querySelector('header')`)
- Birth date as calendar date: `babies.birth_date` now stored/sent as pure local `YYYY-MM-DD` (frontend stops `toISOString()`, backend `normalizeBirthDate` converts legacy RFC3339 via `X-Timezone-Offset`); readers use new `parseLocalDate` util — date-only string kept local, RFC3339 (legacy rows) converted to viewer-local; growth CSV export keeps 成长 rows date-only instead of `In(user zone)`

### Known Issues
- `vue-tsc` typecheck fails on Node.js v24 — not a code issue
- No `.gitignore` exists; `baby-care-tracker.exe` binary is tracked in git

## Architecture
- **Backend**: Go + gin on `:8080`, modernc.org/sqlite, WebSocket broadcast via Hub pattern
- **Frontend**: Vue 3 + Pinia + Vue Router, Vite dev server on `:5173`
- **DB**: `C:\app\data\app.db`; DATETIME columns always return UTC with `Z` suffix

## Key Decisions
- Timezone: client offset header, not server TZ; stored times are UTC
- Calendar dates (birth date, growth measured date) are pure local `YYYY-MM-DD`, never instants — no timezone conversion on store/read
- Record events: `record-created` and `record-deleted` CustomEvents on `window` — WebSocket broadcasts trigger these; delete handlers dispatch directly for instant UI
