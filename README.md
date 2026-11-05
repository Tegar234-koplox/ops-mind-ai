# OpsMind AI

A manufacturing operations investigation platform: deterministic Go detection → a Meta agent selects cross-department evidence → report → human review. The entire dataset is synthetic. Four P0 rules and four optional P1 rules, eight read tools, persistent PostgreSQL, React/TypeScript, login, SSE and audit.

## Live demo

[OpsMind AI — login-protected demo](https://opsmind-ai-l0px.onrender.com). Credentials are provided privately by the owner. All manufacturing records are synthetic. The Free service may have cold starts.

Verified on 29 October 2026: P0, the four P1 rules, hover/focus and the shift handover report are running online. The real Meta golden investigation (two rounds, seven tools, four domains), evidence, SSE reconnect and human review succeeded; reports stayed intact after deployment. Building a handover report adds no model calls. See the [evidence matrix and claim limits](docs/RELEASE_READINESS.md). The video and submission are handled by the owner and have not been declared complete.

## Run locally with Docker

Prerequisites: Docker Engine/Compose running. For development without containers: Go 1.27.3, Node 24, PostgreSQL 18.

1. Copy `.env.example` to `.env`.
2. Set `OPERATOR_PASSWORD` to at least 12 characters and `SESSION_SECRET` to at least 32 random characters. Do not use an example password in a deployment. Alternative: a bcrypt `OPERATOR_PASSWORD_HASH` from `OPERATOR_PASSWORD=... go run ./cmd/opsmind hash-password`.
3. Set `MODEL_API_KEY` for Meta and keep `MODEL_MODE=meta`. For testing, select `MODEL_MODE=mock` explicitly. There is no silent fallback.
4. Run `docker compose up --build`.
5. Open http://localhost:8080 and log in; press **Run detection**. OT-001 automatically goes into investigation.

Compose uses a local database on port 55432, the frontend/API on port 8080, and seeds the database when it is empty. Do not run the portable database and Compose on the same port. `APP_ORIGIN` must match the browser origin; locally `COOKIE_SECURE=false`, over HTTPS `true`.

## Without Docker on Windows

Portable Go in `.tools/go` can be used if it was set up during a development session. The portable PostgreSQL package is only a local testing tool, not a production dependency:

```powershell
npm.cmd install --prefix .tools/pg embedded-postgres@18.4.0-beta.17
node scripts/portable-postgres.mjs
```

In another terminal, set `DATABASE_URL=postgres://opsmind:local-only-database@127.0.0.1:55432/opsmind?sslmode=disable` in `.env`, then:

```powershell
Set-Location web
npm.cmd ci
npm.cmd run build
Set-Location ..
.\\scripts\\run.ps1 -Mock
```

`-Mock` explicitly selects simulation for that process; the `.env` file can still use the Meta default. If the local configuration was generated automatically, the random password is available only in `.env`. Do not upload that file. For real Meta calls, run without `-Mock` after filling in the key. CLI: `serve`, `migrate`, `seed`, `seed-p1`, `detect`, `hash-password`. The CLI reads the environment; the PowerShell scripts load `.env` without evaluating its contents as code.

## Golden scenario

Assembly Line A, S1, 26 October 2026: overtime cost Rp1,380,000 against a Rp1,000,000 baseline (+38%); output 830/1000 (-17%); 38 assignments = 31 verified present + 3 approved leave + 4 pending; union downtime 110 minutes. Seed 20261027, 150 workers, 3 lines, 8 machines, 60 calendar dates. Fixture evaluation time: 27 October 2026 06:15 WIB; this is not today's live data. Ground truth lives in `eval/` and is not read by the runtime or the model.

## Four P1 rules

P1 was added at the owner's instruction after P0 was stable. Enable it with `ENABLE_P1_RULES=true`; the default stays false. For additional sample data, run `seed-p1` or set `SEED_P1_DEMO=true` explicitly. The fixture only extends the recognized synthetic dataset and does not reset the golden report. ATT-001 requires an absence confirmation, OT-002 reads approval history, and COST-001/COST-002 require an activity baseline and posting closure. Cost threshold: `COST_VARIANCE_BPS=2000`.

Use `gofmt` for Go and `npm run format`/`npm run format:check` in `web` to format TypeScript. The P1 rules are kept separate from the main detection flow; migrations, queries, seed and UI stay in the same monolith.

## Shift handover report (P2)

Open **Shift handover**, choose a line, date and completed shift, then **Create new report**. The report contains four-domain facts from Go, notes that need verification, machine disruptions, cases, and saved investigation/review results. Every version is immutable; later review changes do not alter older reports. Use the version picker to open the archive and **Print / Save PDF** to open the browser print dialog. On phones, use the workspace picker button at the top of the page.

Building a report adds no Meta calls and does not change operational data. The operational evaluation time still follows the synthetic fixture; review status is captured when the report is created. A snapshot is not a shift acceptance or an approval of actions. Source references and hashes support traceability; they do not guarantee that every narrative is correct. The feature uses migration 000005; do not run the down migration on a database whose reports must be kept.

## Testing the application

```powershell
. .\\scripts\\load-env.ps1
$env:TEST_DATABASE_URL='postgres://opsmind:local-only-database@127.0.0.1:55432/opsmind_test?sslmode=disable'
& $GoCommand test -count=1 ./...
& $GoCommand vet ./...
& $GoCommand build -o bin/opsmind.exe ./cmd/opsmind
Set-Location web
npm.cmd ci
npm.cmd run build
npx.cmd playwright install chromium
npm.cmd run test:e2e
```

Integration tests **reset the schema only on the TEST_DATABASE_URL database whose name ends in _test**. Never point it at a business database. Without that variable, integration tests are skipped and must be reported as skipped. Linux: `go test -race ./...` with PostgreSQL 18; CI provisions both and uses the mock, with no Meta key. Browser tests use the operator from the environment or `.env`; do not enable the mock in the public application without a label.

## Free deployment

Use `render.yaml`: one Render Free Docker web service and a Neon Free PostgreSQL 18 database with TLS and a direct connection (the worker's session lock needs a persistent PostgreSQL session). Use `sslmode=verify-full`. The Blueprint uses the Singapore region and the main branch, and deploys after CI passes. Provide DATABASE_URL, APP_ORIGIN (the exact deployment URL), OPERATOR_PASSWORD_HASH and MODEL_API_KEY as secret environment variables. SESSION_SECRET is generated by Render. BUDGET_PHASE=demo, COOKIE_SECURE=true, SEED_IF_EMPTY=true. Do not create paid services or add a payment method without the owner's decision.

Render Free may sleep after 15 minutes without inbound traffic; a cold start takes about one minute. The worker resumes from checkpoints while the process is running. All state is in PostgreSQL; the container filesystem is ephemeral. This is a demo built with a production approach, not a service with an SLA. Application limit US$20 (15 development + 5 demo), at most US$0.25 per run; the full US$150 credit may not stay available forever. Meta input/output pricing was verified during planning and must be rechecked before release.

## Security and product boundaries

The model has no free-form SQL and no write tools. Every tool is scoped to a shift or case. Numeric facts come from Go, hypotheses from the model. A valid hash, reference or schema does not guarantee that the narrative or the cause and effect is correct. Four pending attendance records do not mean misconduct, and abnormal cost does not prove fraud. Approval only records a decision; no business action is executed.

One operator account from the environment, with an 8-hour signed cookie. Logout removes the browser cookie; the MVP cannot yet revoke all sessions across devices. Idempotency replays the stored response bytes; an operation that crashes before its response is saved stays pending to prevent ambiguous retries. Check the case state before sending a new operation. Reviews of older reports are rejected. The audit log is append-only.

Not yet included: ERP, predictive maintenance, public signup, multitenancy, or an enterprise production system.

## Documents and sources

The ten planning documents are in `docs/`; the ERD is in DATABASE_DESIGN and the component diagram in ARCHITECTURE. HACKATHON_DEMO_PLAN contains the 180-second script and the submission checklist. Internal deadline 4 November 2026 18:00 WIB; event 5 November 09:00 WIB. The public repository, recording, deployment, real Meta calls and submission are each declared complete only after they have been verified.

- https://dev.meta.ai/events/global-hackathon
- https://dev.meta.ai/docs/tool-calling
- https://dev.meta.ai/docs/structured-output
- https://dev.meta.ai/docs/pricing-rate-limits
- https://render.com/docs/free
- https://neon.com/pricing

The application code uses the [MIT license](LICENSE); libraries and icons keep their [third-party licenses and attributions](THIRD_PARTY_NOTICES.md). Manufacturing data is synthetic. Codex was used as a development aid, while the application's reasoning engine uses Meta. The onboarding summary is the basis for the new-work requirement; the final Official Rules/Event Waiver and the assistant terms still need to be checked. There is no claim of explicit organizer approval for Codex. See the [source status and submission checklist](docs/SUBMISSION_REQUIREMENTS.md).

## Where to continue

Read the [handoff notes](docs/HANDOFF.md) before continuing work. Test status, publication constraints and scope limits are recorded there.

## Indonesian / English

Choose Indonesia or English with the language control on the login page or in the workspace header. The preference is stored in the browser; if browser storage is unavailable, the selector still works for that page session. Labels, statuses, error messages, facts and number/time formats follow the selected language, while the operational time zone stays Asia/Bangkok (WIB).

The language of new AI reports follows the selection at the moment an investigation starts, including automatic investigations from new detections and follow-up investigations. Changing the display language does not change the language of a run in progress and does not call the model. Older reports/syntheses and review notes stay in their original language, with a report language marker. Standard handover labels and notes are translated for display; archived payloads and hashes are not rewritten. For a new report in another language on an older case, use a follow-up investigation in the chosen language.

Migration 000006 adds report_language; API clients that do not send that property still default to `id` (Indonesian). There is no translation service and no additional frontend dependency.
