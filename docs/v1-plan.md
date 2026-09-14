# Sabeel V1 Product Plan

Last updated: 2026-09-14

## Purpose

V1 makes Sabeel a complete social reading product for books. A reader can find
the correct edition of a book, build and manage a library, record progress,
write notes and reviews, follow other readers, and return to a useful home
experience that combines reading activity with relevant discovery.

Books are the only first-class source type in V1. The architecture may leave
room for articles, courses, podcasts, and other media later, but V1 does not
ship incomplete experiences for those types.

## Status Legend

| Status | Meaning |
| --- | --- |
| Done | Implemented and usable in the current product. |
| Partial | A usable foundation exists, but it does not meet the V1 exit criteria. |
| Planned | Agreed V1 work that has not been implemented. |
| Later | Deliberately outside V1. |

## Product Decisions

1. **Catalog:** use Open Library as the primary catalog foundation and Google
   Books as a live fallback and metadata gap detector. Keep provider adapters
   behind Sabeel's own catalog model so no provider becomes the product model.
2. **Data acquisition:** prefer official APIs, licensed datasets, publisher
   feeds, and user corrections. Do not scrape Google Books, Goodreads, Amazon,
   or another site whose terms prohibit it.
3. **Book identity:** model a creative work separately from its physical or
   digital editions. Reviews, ratings, shelves, and social activity generally
   belong to the work; ISBN, cover, format, language, and pagination belong to
   an edition.
4. **Home priority:** reading utility comes first, high-signal social activity
   second, and recommendations third. Dedicated Following and Discover views
   make each mode available without forcing one mixed feed to do everything.
5. **Recommendations:** start with transparent rules and collaborative signals.
   Do not block V1 on machine-learning infrastructure.
6. **Clients:** deliver one responsive web application and installable PWA.
   Native mobile applications are later, but V1 APIs and interaction patterns
   must not prevent them.
7. **Identity:** Sabeel owns its Zitadel deployment and configuration. It has no
   runtime, network, secret, or operational dependency on another project.

## Current Baseline

| Capability | Status | Current state | V1 gap |
| --- | --- | --- | --- |
| Zitadel sign-up and sign-in | Done | OIDC Authorization Code + PKCE, local opaque sessions, logout, verified identity linking. | Operate and test Sabeel's own dev/prod Zitadel stack. |
| Profiles | Done | Private profile editing and public profiles. | Add onboarding preferences and social counts. |
| Book records | Partial | Manual source creation, contributors, and basic book metadata. | Work/edition model, identifiers, provenance, deduplication, provider ingestion. |
| Catalog search | Partial | Local title search. | Ranked local search, external fallback, edition selection, import review. |
| Personal library | Done | Reading state, progress, visibility, and removal. | Move associations to works with an optional preferred edition. |
| Notes | Done | Private/public notes on a source. | Preserve behavior through catalog migration and improve book-page presentation. |
| Reviews | Done | One review per user and source, with public views. | Work-level ratings/reviews, aggregate ratings, social interactions. |
| Collections | Partial | Create, update, delete, and publish collections. | Replace JSON source IDs with ordered relational entries. |
| Social graph and activity | Planned | No follows or activity feed. | Follows, privacy rules, high-signal activity events, Following view. |
| Recommendations | Planned | No preference model or recommendation engine. | Onboarding signals, explainable ranking, Discover view. |
| Responsive web UX | Partial | Responsive application shell and core screens exist. | Complete navigation, loading/empty/error states, accessibility, compact/mobile workflows. |
| PWA and offline behavior | Planned | No manifest, service worker, or sync strategy. | Installability, safe asset/read caching, offline drafts, recovery. |
| Operations | Partial | Containers, migrations, sqlc, and automated tests exist. | Independent identity stack, production Compose path, backups, observability, deployment checks. |

## V1 Experience

### First Run

1. A visitor signs up or signs in through Sabeel's branded Zitadel experience.
2. Sabeel asks for a display name and a small set of optional reading interests.
3. The reader adds at least three books or follows people before recommendation
   personalization is expected to become useful.
4. The reader lands on Home with actionable library context, social activity
   when available, and discovery that clearly explains why it is shown.

### Home And Feeds

**Home** is the default and uses this precedence:

1. Continue reading, recent library items, unfinished drafts, and reader goals.
2. High-signal activity from followed people: reviews, ratings with commentary,
   books started or finished, and curated collections.
3. A compact set of recommendations with explanations such as “because you
   liked…” or “popular among readers you follow.”

**Following** is chronological and social. It omits noisy events such as every
progress increment, profile edit, or collection reorder.

**Discover** is recommendation-heavy and supports browsing by subjects,
contributors, community interest, and editorial collections.

When a new user has no graph or library history, Home uses onboarding interests,
globally useful lists, and clear prompts to add books or follow readers.

## Workstreams

### 1. Catalog And Data Pipeline

Status: **Partial**

Deliverables:

- Introduce `book_works`, `book_editions`, contributors, contributor roles,
  external identifiers, subjects, images, and field-level metadata provenance.
- Migrate existing source records without losing library items, notes, reviews,
  collection membership, or public URLs.
- Add provider interfaces for search, lookup, normalization, and cover retrieval.
- Import an initial Open Library dataset appropriate for V1 and support
  incremental refreshes.
- Use Google Books only as an on-demand fallback or enrichment source, with
  attribution, caching, quotas, and provider terms respected.
- Normalize ISBNs and provider IDs; merge conservatively and keep an auditable
  record of source values and merge decisions.
- Let users propose metadata corrections without directly overwriting trusted
  provider values.
- Add ingestion metrics, dead-letter handling, retry policy, and repeatable jobs.

Exit criteria:

- Searching common books returns useful work-level results and available
  editions without creating obvious duplicates.
- ISBN lookup selects the corresponding edition and links it to the correct work.
- Provider outages degrade to the local catalog; they do not break book pages.
- Every imported field can be traced to a provider, import time, and confidence.
- Re-running an import is idempotent and covered by fixture-based tests.

### 2. Core Reading Experience

Status: **Partial**

Deliverables:

- Build work and edition detail pages with contributors, subjects, description,
  cover, publication data, aggregate community data, and edition switching.
- Add fast library actions for Want to Read, Reading, Read, and Did Not Finish.
- Preserve progress, visibility, notes, reviews, and collections during the
  catalog migration.
- Replace collection `source_ids` JSON with ordered collection-entry rows.
- Add robust empty, loading, partial-data, duplicate, and provider-error states.

Exit criteria:

- A reader can search, inspect an edition, add the work to their library, update
  progress, review it, add notes, and place it in an ordered collection.
- Work-level content remains stable when a preferred edition changes.
- All public/private visibility rules have service and handler coverage.

### 3. Social Graph And Activity

Status: **Planned**

Deliverables:

- Follow and unfollow public profiles, with follower/following lists and counts.
- Define durable activity events for reviews, annotated ratings, reading-state
  transitions, and published collections.
- Build a paginated Following feed with deterministic ordering and privacy
  filtering at query time.
- Support likes and comments on reviews only if moderation and notification
  requirements fit the V1 schedule; otherwise defer both together.
- Add block/report controls before broad public launch.

Exit criteria:

- Activity never reveals a private profile, library item, note, or collection.
- Feed generation is idempotent and does not duplicate events on retries.
- A user can understand and control why another account appears in their feed.

### 4. Recommendations And Discovery

Status: **Planned**

Deliverables:

- Capture optional onboarding interests and use library behavior as implicit
  signals with clear privacy boundaries.
- Build a rule-based ranker using subject affinity, contributor affinity,
  community popularity, recency, and followed-reader overlap.
- Filter books already read or dismissed and diversify repeated contributors and
  subjects.
- Store recommendation reason codes so every result can be explained.
- Measure impressions, opens, saves, dismissals, and subsequent reading actions.

Exit criteria:

- Home and Discover produce deterministic, explainable results for seeded users.
- Cold-start users receive useful non-personalized discovery.
- Users can dismiss a recommendation and it does not immediately return.
- Recommendation work can evolve behind a stable interface without changing the
  catalog or client contracts.

### 5. Responsive UX And PWA

Status: **Partial**

Deliverables:

- Finish responsive app navigation for compact mobile, tablet, and desktop
  layouts, with Home, Search, Library, Following, Discover, and Profile.
- Meet keyboard, focus, contrast, reduced-motion, and screen-reader expectations.
- Add a web app manifest, icons, install metadata, and a versioned service worker.
- Cache the application shell and safe public/read responses; never cache auth
  callbacks, session responses, or private API data in shared caches.
- Store note/review drafts locally with an explicit retry and conflict experience.
- Validate layout and core workflows with automated browser tests at mobile and
  desktop viewports.

Exit criteria:

- The complete critical path works at 360 px width and on current desktop
  browsers without overlap, horizontal overflow, or inaccessible controls.
- The application is installable and opens to a useful offline state.
- Losing connectivity while writing does not lose the draft.
- Authentication and private data do not leak through PWA caches.

### 6. Platform, Security, And Operations

Status: **Partial**

Deliverables:

- Run Sabeel, its application database, and its Zitadel database/services on
  Sabeel-owned Compose networks and volumes.
- Automate creation of the Sabeel Zitadel project and confidential web client.
- Provide separate development and production configurations with no dependency
  on a sibling repository or shared container network.
- Require HTTPS, secure cookies, durable secret management, SMTP, backups, and
  an offline break-glass administrator for production.
- Add structured logs, request IDs, service health, error reporting, metrics,
  database and identity backups, and a restore drill.
- Add abuse controls for auth-adjacent and write-heavy endpoints.

Exit criteria:

- A clean checkout can start the complete development stack using only files in
  this repository.
- A production host can deploy from documented Sabeel artifacts and secrets
  without another project being present.
- Sign-up, email verification, sign-in, callback, session renewal, and logout are
  exercised end to end before release.
- Backup restoration is tested for both Sabeel and Zitadel databases.

## Delivery Sequence

| Milestone | Status | Outcome |
| --- | --- | --- |
| M0: Independent foundation | In progress | Sabeel-owned dev/prod Zitadel, repeatable startup, auth verification, and this V1 baseline. |
| M1: Catalog core | Planned | Work/edition schema, migration, provider contracts, Open Library import, Google fallback, provenance. |
| M2: Complete book loop | Planned | Search, work/edition pages, library, notes, reviews, relational collections, polished responsive flows. |
| M3: Social reading | Planned | Follows, privacy-safe activity model, Following feed, profile social surfaces, launch moderation minimums. |
| M4: Home and discovery | Planned | Blended Home, Discover, onboarding interests, explainable heuristic recommendations, analytics. |
| M5: PWA and launch hardening | Planned | Installability, offline drafts, browser coverage, accessibility, performance, monitoring, backups, release checks. |

Catalog identity and migration land before feed or recommendation work because
both features must refer to stable work IDs. The responsive app shell can proceed
in parallel once those page and navigation contracts are defined.

## Explicitly Outside V1

- Native iOS or Android applications.
- Media types other than books.
- Machine-learned embeddings or model training as a recommendation dependency.
- Direct messaging, groups, clubs, events, or real-time chat.
- Retail purchasing, marketplace features, or subscriptions.
- Full publisher ONIX ingestion unless it becomes necessary to fill a validated
  catalog gap during M1.
- Broad scraping of commercial book or social-reading websites.

## Cross-Cutting Release Requirements

- Database changes are reversible where practical and have a tested data
  migration path.
- Public/private authorization is enforced server-side and covered by tests.
- User-generated content has report, block, retention, and deletion policies.
- Product analytics avoid storing note/review bodies or identity secrets.
- Core pages have meaningful loading, empty, partial, and failure states.
- CI runs Go tests, sqlc generation/vetting, frontend checks, production builds,
  and critical browser flows.
- Operational documentation includes deploy, rollback, backup, restore, secret
  rotation, and identity recovery procedures.

## Definition Of V1 Done

V1 is complete when a new reader can independently register, verify their
identity, discover the correct book and edition, manage a reading library, write
and publish notes/reviews, organize books, follow people, consume a useful Home
and Following experience, receive explainable recommendations, and use the
critical workflow on mobile-sized and desktop screens as an installable PWA.

The release must run from Sabeel-owned development and production infrastructure,
meet the security and privacy criteria above, and have demonstrated backup and
restore procedures. A feature being visible in the UI is not sufficient: its
failure, privacy, migration, and operational paths must also be ready.

## Immediate Next Actions

1. Finish M0 and record a passing clean-stack authentication exercise.
2. Write the catalog architecture decision record and migration plan.
3. Implement work/edition schema and provider-neutral catalog interfaces.
4. Build a small Open Library fixture import and Google Books fallback spike.
5. Validate search quality and merge rules against a representative book set.
6. Lock the responsive information architecture before M2 page implementation.
