import {
  Book,
  CheckCircle2,
  Layers3,
  Library,
  Loader2,
  LogOut,
  Plus,
  Star,
  StickyNote,
} from "lucide-react";
import type { FormEvent, ReactNode } from "react";
import { Link } from "react-router";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import type { Collection, LibraryItemWithSource, Note, Review, Source } from "~/lib/api";

export type DashboardData = {
  sources: Source[];
  library: LibraryItemWithSource[];
  notes: Note[];
  reviews: Review[];
  collections: Collection[];
};

export type RunAction = (action: () => Promise<unknown>, fallback: string) => Promise<void>;

export function DashboardHeader({
  displayName,
  onLogout,
}: {
  displayName: string;
  onLogout: () => void;
}) {
  return (
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-emerald-600 p-2">
            <Library className="h-5 w-5 text-white" />
          </div>
          <span className="text-xl font-bold text-slate-900">Sabeel</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="hidden text-sm text-slate-600 sm:inline">Welcome, {displayName}</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={onLogout}
            className="text-slate-600 hover:text-red-600"
          >
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </Button>
        </div>
      </div>
    </header>
  );
}

export function DashboardIntro() {
  return (
    <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 className="text-3xl font-bold text-slate-900">Dashboard</h1>
        <p className="mt-2 text-slate-600">
          Build and test your knowledge library with real platform data.
        </p>
      </div>
      <Link to="/settings">
        <Button variant="outline">Profile settings</Button>
      </Link>
    </div>
  );
}

export function DashboardStats({ data }: { data: DashboardData }) {
  return (
    <div className="mb-8 grid grid-cols-2 gap-4 lg:grid-cols-5">
      <StatCard icon={<Book />} label="Book sources" value={data.sources.length} />
      <StatCard icon={<Library />} label="Library" value={data.library.length} />
      <StatCard icon={<StickyNote />} label="Notes" value={data.notes.length} />
      <StatCard icon={<Star />} label="Reviews" value={data.reviews.length} />
      <StatCard icon={<Layers3 />} label="Collections" value={data.collections.length} />
    </div>
  );
}

export function LibraryPanel({
  items,
  loading,
  onDelete,
  onUpdate,
}: {
  items: LibraryItemWithSource[];
  loading: boolean;
  onDelete: (id: string) => void;
  onUpdate: (id: string, payload: unknown) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Library className="h-5 w-5 text-emerald-600" />
          My Library
        </CardTitle>
        <CardDescription>
          Your tracked sources. New books created here are added to your library automatically.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <LoadingRow />
        ) : items.length === 0 ? (
          <EmptyState message="No library items yet. Create a book to start testing the platform." />
        ) : (
          <div className="space-y-3">
            {items.slice(0, 8).map((item) => (
              <LibraryItemRow key={item.id} item={item} onDelete={onDelete} onUpdate={onUpdate} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function SourcesPanel({
  librarySourceIDs,
  loading,
  onAdd,
  sources,
}: {
  librarySourceIDs: Set<string>;
  loading: boolean;
  onAdd: (sourceID: string) => void;
  sources: Source[];
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Available Book Sources</CardTitle>
        <CardDescription>Public book records currently in the platform.</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <LoadingRow />
        ) : sources.length === 0 ? (
          <EmptyState message="No book sources yet." />
        ) : (
          <div className="space-y-3">
            {sources.slice(0, 8).map((source) => (
              <SourceRow
                key={source.id}
                isInLibrary={librarySourceIDs.has(source.id)}
                onAdd={onAdd}
                source={source}
              />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function AddBookCard({
  author,
  isbn,
  saving,
  title,
  onAuthorChange,
  onISBNChange,
  onSubmit,
  onTitleChange,
}: {
  author: string;
  isbn: string;
  saving: boolean;
  title: string;
  onAuthorChange: (value: string) => void;
  onISBNChange: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onTitleChange: (value: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Plus className="h-5 w-5 text-emerald-600" />
          Add a Book
        </CardTitle>
        <CardDescription>Minimal book creation for MVP testing.</CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={onSubmit}>
          <Input
            value={title}
            onChange={(event) => onTitleChange(event.target.value)}
            placeholder="Book title"
            required
          />
          <Input
            value={author}
            onChange={(event) => onAuthorChange(event.target.value)}
            placeholder="Author"
          />
          <Input
            value={isbn}
            onChange={(event) => onISBNChange(event.target.value)}
            placeholder="ISBN-13"
          />
          <Button type="submit" className="w-full" disabled={saving || !title.trim()}>
            {saving ? "Saving..." : "Create and add"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

export function SourceActivityCard({
  collectionName,
  collectionPublic,
  librarySources,
  noteContent,
  notePublic,
  reviewContent,
  reviewRating,
  saving,
  selectedSourceID,
  onCollectionNameChange,
  onCollectionPublicChange,
  onNoteContentChange,
  onNotePublicChange,
  onReviewContentChange,
  onReviewRatingChange,
  onSelectedSourceChange,
  onSubmitCollection,
  onSubmitNote,
  onSubmitReview,
}: {
  collectionName: string;
  collectionPublic: boolean;
  librarySources: Array<{ item: LibraryItemWithSource; source: LibraryItemWithSource["source"] }>;
  noteContent: string;
  notePublic: boolean;
  reviewContent: string;
  reviewRating: string;
  saving: boolean;
  selectedSourceID: string;
  onCollectionNameChange: (value: string) => void;
  onCollectionPublicChange: (value: boolean) => void;
  onNoteContentChange: (value: string) => void;
  onNotePublicChange: (value: boolean) => void;
  onReviewContentChange: (value: string) => void;
  onReviewRatingChange: (value: string) => void;
  onSelectedSourceChange: (value: string) => void;
  onSubmitCollection: (event: FormEvent<HTMLFormElement>) => void;
  onSubmitNote: (event: FormEvent<HTMLFormElement>) => void;
  onSubmitReview: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Work With a Source</CardTitle>
        <CardDescription>
          Select a library item, then add notes, reviews, and collections.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <select
          className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm"
          value={selectedSourceID}
          onChange={(event) => onSelectedSourceChange(event.target.value)}
          disabled={librarySources.length === 0}
        >
          {librarySources.length === 0 ? (
            <option value="">No library sources yet</option>
          ) : (
            librarySources.map(({ item, source }) => (
              <option key={item.id} value={item.source_id}>
                {source?.title || item.source_id}
              </option>
            ))
          )}
        </select>

        <form className="space-y-3" onSubmit={onSubmitNote}>
          <textarea
            className="min-h-24 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm"
            value={noteContent}
            onChange={(event) => onNoteContentChange(event.target.value)}
            placeholder="Write a note..."
          />
          <label className="flex items-center gap-2 text-sm text-slate-600">
            <input
              type="checkbox"
              checked={notePublic}
              onChange={(event) => onNotePublicChange(event.target.checked)}
            />
            Public note
          </label>
          <Button
            type="submit"
            variant="outline"
            className="w-full"
            disabled={saving || !selectedSourceID || !noteContent.trim()}
          >
            Add note
          </Button>
        </form>

        <form className="space-y-3 border-t border-slate-200 pt-4" onSubmit={onSubmitReview}>
          <div className="grid grid-cols-[90px_1fr] gap-3">
            <select
              className="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm"
              value={reviewRating}
              onChange={(event) => onReviewRatingChange(event.target.value)}
            >
              {[5, 4, 3, 2, 1].map((rating) => (
                <option key={rating} value={rating}>
                  {rating} star
                </option>
              ))}
            </select>
            <Input
              value={reviewContent}
              onChange={(event) => onReviewContentChange(event.target.value)}
              placeholder="Short review"
            />
          </div>
          <Button
            type="submit"
            variant="outline"
            className="w-full"
            disabled={saving || !selectedSourceID}
          >
            Add public review
          </Button>
        </form>

        <form className="space-y-3 border-t border-slate-200 pt-4" onSubmit={onSubmitCollection}>
          <Input
            value={collectionName}
            onChange={(event) => onCollectionNameChange(event.target.value)}
            placeholder="Collection name"
          />
          <label className="flex items-center gap-2 text-sm text-slate-600">
            <input
              type="checkbox"
              checked={collectionPublic}
              onChange={(event) => onCollectionPublicChange(event.target.checked)}
            />
            Public collection
          </label>
          <Button
            type="submit"
            variant="outline"
            className="w-full"
            disabled={saving || !selectedSourceID || !collectionName.trim()}
          >
            Create collection
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

export function RecentNotesCard({
  notes,
  onDelete,
}: {
  notes: Note[];
  onDelete: (id: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Notes</CardTitle>
      </CardHeader>
      <CardContent>
        {notes.length === 0 ? (
          <EmptyState message="No notes yet." />
        ) : (
          <div className="space-y-3">
            {notes.slice(0, 5).map((note) => (
              <div
                key={note.id}
                className="rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-700"
              >
                <p>{note.content}</p>
                <Button
                  variant="ghost"
                  size="xs"
                  className="mt-2 text-red-600"
                  onClick={() => onDelete(note.id)}
                >
                  Delete
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function RecentReviewsCard({
  reviews,
  onDelete,
}: {
  reviews: Review[];
  onDelete: (id: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Reviews</CardTitle>
      </CardHeader>
      <CardContent>
        {reviews.length === 0 ? (
          <EmptyState message="No reviews yet." />
        ) : (
          <div className="space-y-3">
            {reviews.slice(0, 5).map((review) => (
              <div
                key={review.id}
                className="rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-700"
              >
                <p className="font-medium">{review.rating}/5 stars</p>
                {review.content && <p className="mt-1">{review.content}</p>}
                <Button
                  variant="ghost"
                  size="xs"
                  className="mt-2 text-red-600"
                  onClick={() => onDelete(review.id)}
                >
                  Delete
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function CollectionsCard({
  collections,
  onDelete,
}: {
  collections: Collection[];
  onDelete: (id: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Collections</CardTitle>
      </CardHeader>
      <CardContent>
        {collections.length === 0 ? (
          <EmptyState message="No collections yet." />
        ) : (
          <div className="space-y-3">
            {collections.slice(0, 5).map((collection) => (
              <div
                key={collection.id}
                className="rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-700"
              >
                <p className="font-medium">{collection.name}</p>
                <p className="mt-1 text-slate-500">
                  {collection.is_public ? "Public" : "Private"} ·{" "}
                  {collection.source_ids?.length || 0} sources
                </p>
                <Button
                  variant="ghost"
                  size="xs"
                  className="mt-2 text-red-600"
                  onClick={() => onDelete(collection.id)}
                >
                  Delete
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function LibraryItemRow({
  item,
  onDelete,
  onUpdate,
}: {
  item: LibraryItemWithSource;
  onDelete: (id: string) => void;
  onUpdate: (id: string, payload: unknown) => void;
}) {
  const source = item.source;
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <Link
            to={`/sources/books/${item.source_id}`}
            className="font-medium text-slate-900 hover:text-emerald-700"
          >
            {source?.title || item.source_id}
          </Link>
          <p className="mt-1 text-sm text-slate-500">
            {item.status.replace("_", " ")} · {item.visibility}
          </p>
        </div>
        <div className="flex flex-col items-end gap-2">
          <CheckCircle2 className="h-5 w-5 text-emerald-600" />
          <select
            className="h-8 rounded-md border border-slate-300 bg-white px-2 text-xs"
            value={item.status}
            onChange={(event) => onUpdate(item.id, { status: event.target.value })}
          >
            <option value="to_consume">To consume</option>
            <option value="in_progress">In progress</option>
            <option value="completed">Completed</option>
            <option value="paused">Paused</option>
            <option value="abandoned">Abandoned</option>
          </select>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              size="xs"
              onClick={() =>
                onUpdate(item.id, {
                  visibility: item.visibility === "public" ? "private" : "public",
                })
              }
            >
              {item.visibility === "public" ? "Make private" : "Make public"}
            </Button>
            <Button
              variant="ghost"
              size="xs"
              className="text-red-600"
              onClick={() => onDelete(item.id)}
            >
              Remove
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

function SourceRow({
  isInLibrary,
  onAdd,
  source,
}: {
  isInLibrary: boolean;
  onAdd: (id: string) => void;
  source: Source;
}) {
  return (
    <div className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <Link
          to={`/sources/books/${source.id}`}
          className="font-medium text-slate-900 hover:text-emerald-700"
        >
          {source.title}
        </Link>
        <p className="mt-1 text-sm text-slate-500">{source.publisher || source.type}</p>
      </div>
      {!isInLibrary && (
        <Button variant="outline" size="sm" onClick={() => onAdd(source.id)}>
          Add to library
        </Button>
      )}
    </div>
  );
}

function StatCard({ icon, label, value }: { icon: ReactNode; label: string; value: number }) {
  return (
    <Card className="gap-2 py-4">
      <CardContent className="flex items-center gap-3 px-4">
        <div className="rounded-lg bg-emerald-50 p-2 text-emerald-700">{icon}</div>
        <div>
          <p className="text-2xl font-bold text-slate-900">{value}</p>
          <p className="text-xs text-slate-500">{label}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function LoadingRow() {
  return (
    <div className="flex items-center gap-2 py-8 text-sm text-slate-500">
      <Loader2 className="h-4 w-4 animate-spin" />
      Loading platform data...
    </div>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-6 text-center text-sm text-slate-500">
      {message}
    </div>
  );
}
