# Catalog Foundation Design

Status: Proposed

Last updated: 2026-09-14

## Summary

Sabeel will own a canonical, provider-neutral book catalog. Open Library is the
primary bibliographic source, while Google Books is an on-demand fallback and
metadata gap detector. Provider records are inputs to Sabeel's catalog; their
schemas and availability do not define the product model or sit on the critical
path for an existing book page.

The catalog distinguishes a creative work from its editions. Reader activity
such as library status, reviews, ratings, notes, collections, and social events
belongs to a stable work. ISBNs, publishers, publication dates, languages,
formats, pagination, and edition-specific covers belong to editions.

The migration is additive. Existing book `sources.id` values remain the stable
work identifiers so current user data and public URLs do not need to be
rewritten in one high-risk operation.

## Goals

- Return useful work-level search results with the relevant editions attached.
- Resolve an ISBN to one edition and its canonical work.
- Preserve existing libraries, notes, reviews, collections, and URLs.
- Keep provider-specific data behind a Sabeel-owned domain model.
- Record where imported metadata came from and why a canonical value was used.
- Make imports repeatable, idempotent, observable, and safe to retry.
- Continue serving local book pages when an external provider is unavailable.
- Support English and Arabic catalog data without assuming one display language.
- Provide stable contracts for future feeds, recommendations, and mobile clients.

## Non-Goals

- Importing every Open Library record before V1 can launch.
- Scraping commercial book or social-reading websites.
- Republishing reviews written by other platforms' users.
- Building machine-learned search or recommendation infrastructure.
- Supporting non-book media as first-class V1 catalog entries.
- Building a general-purpose library cataloging or MARC management system.
- Automatically merging ambiguous works solely from fuzzy title similarity.

## Domain Model

### Work

A work represents the intellectual creation readers discuss and track. It owns:

- canonical and alternate titles;
- descriptions and first-publication information;
- original language when known;
- contributors and their work-level roles;
- subjects;
- aggregate Sabeel ratings and activity; and
- links to all known editions.

`sources.id` remains the work identifier during V1. A `book_works` row extends
each book source with work-specific fields. Existing foreign keys continue to
refer to the same UUID.

### Edition

An edition represents a publication or release of a work. It owns:

- ISBN-10 and ISBN-13;
- edition title and subtitle where they differ from the work;
- language, format, publisher, and publication date;
- page count and edition statement;
- cover assets; and
- provider edition identifiers.

A library item may store an optional preferred edition. Changing that edition
does not move or duplicate the reader's review, notes, reading state, or social
activity.

### Contributor

Contributors are reusable entities connected to works or editions through
ordered credits. Roles include author, editor, translator, illustrator, and
narrator. Provider contributor IDs are stored independently so equal names do
not force an automatic identity merge.

### Subject

Subjects are normalized Sabeel concepts with provider aliases. Raw provider
labels remain available for provenance, but discovery and recommendations use
the normalized subject IDs.

## Proposed Storage

The exact column names may be refined in migrations, but the ownership and
relationships below are the contract.

### Canonical Tables

| Table | Purpose |
| --- | --- |
| `book_works` | Work-level metadata extending a book `source`. |
| `book_editions` | Edition metadata linked to one canonical work. |
| `book_identifiers` | Normalized ISBN, Open Library, Google, OCLC, and LCCN identifiers. |
| `contributors` | Canonical contributor identity and display metadata. |
| `book_credits` | Ordered work- or edition-level contributor roles. |
| `subjects` | Normalized Sabeel subjects. |
| `book_subjects` | Work-to-subject assignments with confidence. |
| `book_titles` | Alternate and translated titles with language and title type. |
| `book_redirects` | Duplicate work IDs redirected to a surviving work. |
| `collection_entries` | Ordered relational replacement for collection JSON IDs. |

### Acquisition Tables

| Table | Purpose |
| --- | --- |
| `catalog_records` | Provider record identity, raw payload, hash, and retrieval time. |
| `catalog_mappings` | Provider work/edition/contributor IDs mapped to canonical IDs. |
| `metadata_provenance` | Provider record and confidence supporting a canonical field. |
| `catalog_jobs` | Import or refresh job lifecycle and counters. |
| `catalog_failures` | Retryable and terminal record failures with diagnostic context. |
| `import_batches` | User and catalog imports that can be audited or reversed. |

Canonical values remain typed columns rather than an entity-attribute-value
model. Raw payloads support debugging and future re-normalization, while field
provenance explains selected values without making ordinary reads expensive.

Identifiers are normalized before storage. ISBNs contain digits only, ISBN-10
checksums are validated, and valid ISBN-10 values are converted to ISBN-13 for
matching. Unique constraints apply to identifier type and normalized value at
the appropriate work or edition scope.

## Provider Boundary

Providers implement a small internal interface:

```go
type Provider interface {
	Search(ctx context.Context, query SearchQuery) ([]Candidate, error)
	LookupISBN(ctx context.Context, isbn string) ([]Candidate, error)
	GetWork(ctx context.Context, id ProviderID) (Record, error)
	GetEdition(ctx context.Context, id ProviderID) (Record, error)
}
```

Provider adapters return normalized candidates plus their raw records. They do
not write canonical tables directly. A catalog resolver owns validation,
matching, merge decisions, provenance, and transactional persistence.

Every outbound provider request must use:

- a timeout and cancellation;
- bounded retries with jitter for transient failures;
- provider-specific rate limiting;
- a Sabeel user agent and contact address where required;
- structured request metrics without logging credentials; and
- a cache policy compatible with the provider's terms.

### Open Library

Open Library is the primary source for search, work and edition relationships,
ISBN lookup, contributors, subjects, and covers. Interactive traffic uses its
APIs. Bulk catalog growth uses its published data dumps instead of issuing
large numbers of API requests.

V1 begins with lazy materialization and a curated catalog subset rather than a
complete dump:

1. books selected by readers;
2. books needed by onboarding and editorial collections;
3. representative English and Arabic works; and
4. popular works needed for useful cold-start discovery.

Periodic refreshes compare provider revision or payload hashes and process only
changed records. A provider outage disables enrichment but does not make an
existing Sabeel book unavailable.

### Google Books

Google Books is an on-demand fallback for weak Open Library results, ISBN gaps,
and selected missing metadata. Google results remain visibly attributable and
link to Google Books where its terms require it.

Live Google result sets must not be silently blended, reordered, or presented
as Sabeel-owned search results. Any retained value records its Google volume ID,
retrieval time, attribution requirements, and provenance. The adapter and UI
must be reviewable independently if Google's terms change.

## Search And Materialization

The request path is local-first:

1. Normalize the query and search Sabeel's local catalog.
2. Rank work-level results using title, alternate title, contributor, ISBN, and
   subject matches.
3. Return strong local matches immediately.
4. When local coverage is weak, query Open Library within a strict time budget.
5. Query Google Books only when the Open Library fallback is insufficient.
6. Present external candidates without creating permanent canonical records.
7. Materialize the chosen work and edition transactionally when the reader
   opens, saves, reviews, or otherwise acts on it.
8. Queue non-critical enrichment after the canonical record exists.

PostgreSQL full-text search and trigram similarity are sufficient for V1.
Embeddings and a separate search cluster are not required. Ranking should favor
exact ISBN, exact normalized title, prefix title, contributor, subject, and then
fuzzy title matches. Language and popularity may break ties but must not hide an
exact identifier match.

## Matching And Deduplication

Evidence is considered from strongest to weakest:

1. exact normalized ISBN for an edition;
2. exact provider identifier already mapped to a canonical entity;
3. an explicit provider work-to-edition relationship;
4. exact OCLC or LCCN identifier;
5. normalized title, primary contributor, language, and publication date; and
6. fuzzy title and contributor similarity.

The first four signals may merge automatically when they do not conflict. A
high-confidence compound match may suggest a merge. Fuzzy matches alone never
merge automatically.

Conflicting identifiers create a failure or review task rather than silently
overwriting a record. Every merge records the surviving ID, retired ID,
evidence, actor or job, and timestamp. A `book_redirects` row keeps old URLs and
references resolvable. Merge operations must be reversible until dependent
records have been verified.

## Metadata Selection

Provider values are evidence, not unconditional truth. Selection rules are
field-specific:

- identifiers and explicit work/edition links outrank inferred relationships;
- edition publisher, date, pagination, language, and cover remain edition-level;
- work descriptions prefer complete, language-appropriate values;
- user corrections never overwrite raw provider records;
- trusted editorial corrections may override the canonical value while retaining
  all provenance; and
- a missing value may be enriched, but a lower-confidence empty value never
  clears a known value.

Descriptions and cover images require provider-aware attribution and retention
rules. Sabeel must be able to remove one provider's value and recompute a
canonical record from the remaining evidence.

## Existing Data Migration

The migration proceeds in compatible stages:

1. Add catalog tables without removing or renaming existing columns.
2. Create one `book_works` row for every existing book source, preserving its
   `sources.id` UUID.
3. Create an edition from existing ISBN, publisher, date, language, page count,
   and cover fields when edition evidence exists.
4. Attach existing contributors as ordered work credits.
5. Add an optional `preferred_edition_id` to each library item.
6. Backfill ordered `collection_entries` from each collection's JSON source IDs.
7. Update reads to prefer the new model while dual-writing compatible fields
   during the transition.
8. Validate counts and foreign-key coverage before switching API responses.
9. Stop dual writes and remove deprecated edition fields only in a later,
   independently reversible migration.

Reviews, notes, library items, and social activity continue to reference the
stable work/source UUID. Edition-specific annotations may add an optional
edition reference without changing their work ownership.

Each backfill is idempotent and records its import batch. Deployment supports
mixed application versions while a backfill is running; a schema migration must
not require an immediate all-at-once data rewrite.

## External Reading History And Reviews

Sabeel will not scrape, purchase, or bulk-copy reviews written by users of
Goodreads, StoryGraph, Amazon, or another social platform. External community
review text will not be used to seed Sabeel reviews or its community rating.

Sabeel may import a reader's own data from a file they explicitly provide. V1
should support Goodreads CSV, StoryGraph CSV, and a documented Sabeel CSV
format. Importable fields include:

- ISBN or other book identity;
- shelves and reading state;
- dates read and date added;
- the reader's rating and review text;
- spoiler state and private notes where the export supplies them; and
- ownership or read-count metadata when it maps cleanly to Sabeel.

The import flow is staged:

1. Parse and validate the file into an `import_batch`.
2. Match exact ISBNs and known provider identifiers.
3. Score title and contributor matches for rows without identifiers.
4. Ask the reader to resolve ambiguous or unmatched rows.
5. Preview all creates, updates, and conflicts.
6. Import reviews as private drafts by default.
7. Let the reader explicitly publish selected reviews on Sabeel.
8. Preserve original dates and record `imported_from` without implying the
   external platform authored or endorsed the content.
9. Allow the reader to reverse the batch without deleting later manual edits.

External popularity may be retained as a separately labelled discovery signal
when its license permits it. Sabeel's displayed community rating and review
count include only Sabeel activity.

## API Shape

New APIs use catalog language even while the internal source UUID is preserved:

- `GET /books/search?q=...&language=...`
- `GET /books/works/:workID`
- `GET /books/works/:workID/editions`
- `GET /books/editions/:editionID`
- `GET /books/isbn/:isbn`
- `POST /catalog/materializations`
- `POST /imports/reading-history`
- `GET /imports/:batchID`

Search responses identify local and external candidates, provider attribution,
materialization state, match reason, and the selected display edition. Internal
provider payloads and credentials are never exposed.

Existing `/sources` routes remain compatible during migration and are retired
only after the frontend and automated clients use the book routes.

## Background Processing

Catalog jobs may initially use the existing database and outbox rather than a
new broker. Workers claim jobs with row locking, renew a lease for long jobs,
and make each record operation idempotent.

Job types include:

- provider search-result materialization;
- work and edition enrichment;
- cover retrieval or validation;
- curated and dump-subset imports;
- stale-record refresh;
- canonical metadata recomputation; and
- duplicate review and merge execution.

Failures distinguish retryable provider/network errors from invalid records and
identity conflicts. Terminal failures retain enough context for diagnosis while
excluding secrets and unnecessary personal data.

## Privacy, Security, And Operations

- Provider API keys remain server-side and outside checked-in configuration.
- User import files are private, encrypted in transit, access-controlled, and
  deleted after the configured processing and recovery window.
- Raw catalog payloads are treated as untrusted input and never rendered as HTML.
- Cover URLs and redirects are validated to prevent server-side request forgery.
- Import size, row count, decompression, and request rates are bounded.
- Catalog writes use transactions and database constraints, not only application
  checks, to enforce identity rules.
- Logs include job, provider, record, and request IDs without review bodies or
  user import contents.
- Backups include canonical catalog data, mappings, provenance, and merge history.

Operational metrics include provider latency and errors, cache hit rate, search
fallback rate, unmatched queries, materialization success, duplicate candidates,
job lag, retries, terminal failures, and records missing key metadata.

## Delivery Plan

### Phase 1: Schema And Contracts

- Approve this design and record unresolved policy decisions.
- Add the work, edition, identifier, mapping, provenance, and job tables.
- Define provider-neutral domain types and interfaces.
- Add ISBN and text normalization with fixture tests.

Exit: migrations are reversible, existing behavior is unchanged, and provider
fixtures normalize into stable domain records.

### Phase 2: Existing Data Backfill

- Backfill works and editions from current sources and book metadata.
- Backfill credits and relational collection entries.
- Verify all current user associations and public URLs.

Exit: every existing book has a work, edition data is separated where known,
and before/after association counts match.

### Phase 3: Open Library Integration

- Implement search, ISBN, work, edition, author, and cover operations.
- Add caching, rate limiting, retries, raw records, and provenance.
- Build transactional materialization and duplicate detection.

Exit: fixture and live contract tests can find and persist representative
English and Arabic books without obvious duplicate works.

### Phase 4: Local Search And Google Fallback

- Add ranked local work search and edition-aware results.
- Add bounded Open Library fallback and asynchronous enrichment.
- Add Google Books fallback with the required attribution boundary.

Exit: common title, author, and ISBN searches are useful; provider outages fall
back to local data; weak and unmatched queries are measurable.

### Phase 5: Book Experience And Personal Import

- Move book pages and library actions to work/edition APIs.
- Add edition selection and preferred editions.
- Add previewed, reversible Goodreads/StoryGraph/Sabeel CSV imports.

Exit: a reader can find the correct edition, save the work, review it, organize
it, and safely bring their own reading history into Sabeel.

## Verification Strategy

- Unit tests cover normalization, checksum validation, field precedence,
  matching scores, and provider response mapping.
- Fixture-based provider tests contain successful, partial, malformed, missing,
  multilingual, rate-limited, and conflicting records.
- Repository tests exercise identifier uniqueness, idempotent upserts, redirects,
  provenance, and job claiming under concurrency.
- Migration tests start with representative legacy data and compare every user
  association before and after backfill.
- Contract tests run against provider sandboxes or live read-only endpoints on a
  controlled schedule, not as a requirement for every local test run.
- Browser tests cover local search, fallback, materialization, edition selection,
  library actions, and ambiguous import resolution.
- Failure tests confirm that provider timeouts do not break existing book pages
  and interrupted jobs resume without duplicate records.

## Acceptance Criteria

- Exact ISBN lookup returns the corresponding edition and canonical work.
- Work search does not show obvious edition duplicates as separate works.
- A chosen external candidate becomes a complete local record transactionally.
- Every imported canonical field has provider or editorial provenance.
- Repeating an import produces no duplicate works, editions, or credits.
- Existing reviews, notes, libraries, collections, and URLs survive migration.
- Sabeel pages remain available during provider failures.
- Ambiguous records are held for review instead of merged automatically.
- Imported personal reviews remain private until their owner publishes them.
- Sabeel ratings contain only Sabeel user ratings.

## Open Decisions

- The initial curated languages, subjects, and popularity threshold for bulk
  Open Library ingestion.
- Cover retention and proxying policy for each provider.
- Whether editorial overrides require a dedicated internal administration UI in
  V1 or can begin as an audited command-line workflow.
- The retention period for uploaded personal import files and failed rows.
- Whether imported reading dates support multiple read-throughs in the first
  release of the import feature.

## Provider References

- [Open Library APIs](https://openlibrary.org/developers/api)
- [Open Library work and edition APIs](https://openlibrary.org/dev/docs/api/books)
- [Open Library data dumps](https://openlibrary.org/developers/dumps)
- [Open Library licensing](https://openlibrary.org/developers/licensing)
- [Google Books API overview](https://developers.google.com/books/docs/overview)
- [Google Books branding and attribution](https://developers.google.com/books/branding)
- [Goodreads terms of use](https://www.goodreads.com/about/terms)
- [Goodreads library import and export](https://www.goodreads.com/review/import)
- [StoryGraph user export](https://roadmap.thestorygraph.com/requests-ideas/posts/export-csv)

These are operating dependencies, not merely background reading. Provider
adapters and user-facing attribution must be reviewed against their current
versions before launch and whenever an integration materially changes.
