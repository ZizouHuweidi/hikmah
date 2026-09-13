package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookSeed struct {
	Title       string
	Subtitle    string
	Description string
	ISBN13      string
	Publisher   string
	Author      string
	Tags        []string
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://maktaba:maktaba@localhost:5432/maktaba?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fatal("connect database: %v", err)
	}
	defer db.Close()

	userID := mustUUID()
	if _, err := db.Exec(ctx, `
		INSERT INTO users (id, email, username, password_hash)
		VALUES ($1, 'demo@example.com', 'demo_reader', NULL)
		ON CONFLICT (email) DO UPDATE SET username = EXCLUDED.username
	`, userID.String()); err != nil {
		fatal("seed user: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT id FROM users WHERE email = 'demo@example.com'`).Scan(&userID); err != nil {
		fatal("load user: %v", err)
	}

	profileID := mustUUID()
	if _, err := db.Exec(ctx, `
		INSERT INTO profiles (id, user_id, display_name, bio, public_profile)
		VALUES ($1, $2, 'Demo Reader', 'A public demo profile for exploring Sabeel.', true)
		ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, bio = EXCLUDED.bio, public_profile = true
	`, profileID.String(), userID.String()); err != nil {
		fatal("seed profile: %v", err)
	}

	books := []bookSeed{
		{
			Title:       "The House of Wisdom",
			Subtitle:    "How Arabic Science Saved Ancient Knowledge and Gave Us the Renaissance",
			Description: "A readable account of knowledge transmission through the Islamic Golden Age.",
			ISBN13:      "9780143120568",
			Publisher:   "Penguin",
			Author:      "Jim Al-Khalili",
			Tags:        []string{"history", "science", "islamic-golden-age"},
		},
		{
			Title:       "Lost Enlightenment",
			Subtitle:    "Central Asia's Golden Age from the Arab Conquest to Tamerlane",
			Description: "A broad history of intellectual life in medieval Central Asia.",
			ISBN13:      "9780691165851",
			Publisher:   "Princeton University Press",
			Author:      "S. Frederick Starr",
			Tags:        []string{"history", "central-asia", "knowledge"},
		},
	}

	var sourceIDs []uuid.UUID
	for _, book := range books {
		sourceID := seedBook(ctx, db, book)
		sourceIDs = append(sourceIDs, sourceID)
		seedLibraryAndActivity(ctx, db, userID, sourceID, book.Title)
	}
	seedCollection(ctx, db, userID, sourceIDs)

	fmt.Println("seeded demo data")
	fmt.Println("the demo profile is public; authenticate through Zitadel for private features")
}

func seedBook(ctx context.Context, db *pgxpool.Pool, book bookSeed) uuid.UUID {
	var existingID uuid.UUID
	err := db.QueryRow(ctx, `SELECT id FROM sources WHERE isbn = $1 AND type = 'book' ORDER BY created_at LIMIT 1`, book.ISBN13).Scan(&existingID)
	if err == nil {
		return existingID
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		fatal("load source: %v", err)
	}

	sourceID := mustUUID()
	tags, _ := json.Marshal(book.Tags)
	if _, err := db.Exec(ctx, `
		INSERT INTO sources (id, title, subtitle, type, description, publisher, isbn, tags)
		VALUES ($1, $2, $3, 'book', $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, sourceID.String(), book.Title, book.Subtitle, book.Description, book.Publisher, book.ISBN13, tags); err != nil {
		fatal("seed source: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO book_metadata (source_id, isbn_13, publisher, language)
		VALUES ($1, $2, $3, 'en')
		ON CONFLICT (source_id) DO UPDATE SET isbn_13 = EXCLUDED.isbn_13, publisher = EXCLUDED.publisher
	`, sourceID.String(), book.ISBN13, book.Publisher); err != nil {
		fatal("seed book metadata: %v", err)
	}

	contributorID := mustUUID()
	if err := db.QueryRow(ctx, `
		INSERT INTO contributors (id, name)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, contributorID.String(), book.Author).Scan(&contributorID); err != nil {
		fatal("seed contributor: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO source_contributors (source_id, contributor_id, role, position)
		VALUES ($1, $2, 'author', 0)
		ON CONFLICT DO NOTHING
	`, sourceID.String(), contributorID.String()); err != nil {
		fatal("seed source contributor: %v", err)
	}
	return sourceID
}

func seedLibraryAndActivity(ctx context.Context, db *pgxpool.Pool, userID, sourceID uuid.UUID, title string) {
	if _, err := db.Exec(ctx, `
		INSERT INTO user_library_items (id, user_id, source_id, status, visibility)
		VALUES ($1, $2, $3, 'to_consume', 'public')
		ON CONFLICT (user_id, source_id) DO UPDATE SET visibility = 'public'
	`, mustUUID().String(), userID.String(), sourceID.String()); err != nil {
		fatal("seed library: %v", err)
	}
	noteContent := "A useful starting point for " + title + "."
	if _, err := db.Exec(ctx, `
		DELETE FROM notes WHERE user_id = $1 AND source_id = $2 AND content = $3
	`, userID.String(), sourceID.String(), noteContent); err != nil {
		fatal("reset note: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO notes (id, user_id, source_id, content, content_type, is_public)
		VALUES ($1, $2, $3, $4, 'note', true)
	`, mustUUID().String(), userID.String(), sourceID.String(), noteContent); err != nil {
		fatal("seed note: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO reviews (id, user_id, source_id, rating, content, is_public)
		VALUES ($1, $2, $3, 5, $4, true)
		ON CONFLICT (user_id, source_id) DO NOTHING
	`, mustUUID().String(), userID.String(), sourceID.String(), "Recommended for MVP exploration."); err != nil {
		fatal("seed review: %v", err)
	}
}

func seedCollection(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, sourceIDs []uuid.UUID) {
	encoded, _ := json.Marshal(sourceIDs)
	if _, err := db.Exec(ctx, `DELETE FROM collections WHERE user_id = $1 AND name = 'Demo Reading List'`, userID.String()); err != nil {
		fatal("reset collection: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO collections (id, user_id, name, description, is_public, source_ids)
		VALUES ($1, $2, 'Demo Reading List', 'A seeded public collection.', true, $3)
	`, mustUUID().String(), userID.String(), encoded); err != nil {
		fatal("seed collection: %v", err)
	}
}

func mustUUID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		fatal("uuid: %v", err)
	}
	return id
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
