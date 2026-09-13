import { useQueries, useQuery } from "@tanstack/react-query";
import { BookOpen, Layers3, Library, Loader2, Star, StickyNote } from "lucide-react";
import type { ReactNode } from "react";
import { Link, useParams } from "react-router";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import {
  getPublicProfile,
  listPublicCollectionsByUser,
  listPublicLibrary,
  listPublicNotesByUser,
  listPublicReviewsByUser,
} from "~/lib/api";

export default function PublicProfilePage() {
  const { username = "" } = useParams();
  const profileQuery = useQuery({
    queryKey: ["public-profile", username],
    queryFn: () => getPublicProfile(username),
    enabled: Boolean(username),
  });
  const profile = profileQuery.data;
  const [libraryQuery, notesQuery, reviewsQuery, collectionsQuery] = useQueries({
    queries: [
      {
        queryKey: ["public-library", username],
        queryFn: () => listPublicLibrary(username),
        enabled: Boolean(profile),
      },
      {
        queryKey: ["public-notes", profile?.user_id],
        queryFn: () => listPublicNotesByUser(profile?.user_id || ""),
        enabled: Boolean(profile?.user_id),
      },
      {
        queryKey: ["public-reviews", profile?.user_id],
        queryFn: () => listPublicReviewsByUser(profile?.user_id || ""),
        enabled: Boolean(profile?.user_id),
      },
      {
        queryKey: ["public-collections", profile?.user_id],
        queryFn: () => listPublicCollectionsByUser(profile?.user_id || ""),
        enabled: Boolean(profile?.user_id),
      },
    ],
  });

  const loading =
    profileQuery.isLoading ||
    libraryQuery.isLoading ||
    notesQuery.isLoading ||
    reviewsQuery.isLoading ||
    collectionsQuery.isLoading;
  const hasError =
    profileQuery.isError ||
    libraryQuery.isError ||
    notesQuery.isError ||
    reviewsQuery.isError ||
    collectionsQuery.isError;

  if (loading)
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f6f5ef]">
        <Loader2 className="h-8 w-8 animate-spin text-[#286f62]" />
      </main>
    );
  if (hasError || !profile)
    return (
      <main className="min-h-screen bg-[#f6f5ef] px-4 py-16">
        <div className="mx-auto max-w-2xl rounded-sm border border-[#d7dbd2] bg-[#fbfaf5] p-8 text-center shadow-sm">
          <h1 className="text-2xl font-bold text-[#14221d]">Profile unavailable</h1>
          <p className="mt-3 text-[#67746e]">This profile is private or does not exist.</p>
          <Link to="/">
            <Button className="mt-6">Go home</Button>
          </Link>
        </div>
      </main>
    );

  const library = libraryQuery.data || [];
  const notes = notesQuery.data || [];
  const reviews = reviewsQuery.data || [];
  const collections = collectionsQuery.data || [];

  return (
    <main className="min-h-screen bg-[#f6f5ef] px-4 py-10">
      <div className="mx-auto max-w-6xl">
        <section className="mb-8 rounded-sm border border-[#d7dbd2] bg-[#fbfaf5] p-8 text-[#14221d] shadow-sm">
          <div className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-medium uppercase tracking-wide text-[#1d4236]">
                Public Knowledge Profile
              </p>
              <h1 className="mt-3 text-4xl font-bold">
                {profile.display_name || profile.username || username}
              </h1>
              {profile.bio && <p className="mt-4 max-w-2xl text-[#67746e]">{profile.bio}</p>}
            </div>
            <div className="rounded-sm border border-[#d7dbd2] bg-[#e4ecc9] px-4 py-3 text-sm text-[#1d4236]">
              @{profile.username || username}
            </div>
          </div>
        </section>
        <div className="mb-8 grid grid-cols-2 gap-4 md:grid-cols-4">
          <Stat icon={<Library />} label="Library" value={library.length} />
          <Stat icon={<StickyNote />} label="Notes" value={notes.length} />
          <Stat icon={<Star />} label="Reviews" value={reviews.length} />
          <Stat icon={<Layers3 />} label="Collections" value={collections.length} />
        </div>
        <div className="grid gap-6 lg:grid-cols-2">
          <PublicSection
            title="Public Library"
            description="Sources this reader has chosen to share."
            empty="No public library items yet."
            items={library.map((item) => ({
              id: item.id,
              title: item.source?.title || item.source_id,
              meta: `${item.status.replace("_", " ")} · ${item.visibility}`,
            }))}
          />
          <PublicSection
            title="Public Notes"
            description="Notes and reflections shared by this reader."
            empty="No public notes yet."
            items={notes.map((note) => ({
              id: note.id,
              title: note.content,
              meta: note.content_type,
            }))}
          />
          <PublicSection
            title="Public Reviews"
            description="Ratings and reviews shared publicly."
            empty="No public reviews yet."
            items={reviews.map((review) => ({
              id: review.id,
              title: review.content || `${review.rating}/5 stars`,
              meta: `${review.rating}/5 stars`,
            }))}
          />
          <PublicSection
            title="Public Collections"
            description="Curated lists from this profile."
            empty="No public collections yet."
            items={collections.map((collection) => ({
              id: collection.id,
              title: collection.name,
              meta: `${collection.source_ids?.length || 0} sources`,
            }))}
          />
        </div>
      </div>
    </main>
  );
}

function Stat({ icon, label, value }: { icon: ReactNode; label: string; value: number }) {
  return (
    <Card className="gap-2 py-4">
      <CardContent className="flex items-center gap-3 px-4">
        <div className="rounded-lg bg-[#e4ecc9] p-2 text-[#1d4236]">{icon}</div>
        <div>
          <p className="text-2xl font-bold text-[#14221d]">{value}</p>
          <p className="text-xs text-[#67746e]">{label}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function PublicSection({
  title,
  description,
  empty,
  items,
}: {
  title: string;
  description: string;
  empty: string;
  items: Array<{ id: string; title: string; meta: string }>;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <BookOpen className="h-5 w-5 text-[#286f62]" />
          {title}
        </CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {items.length === 0 ? (
          <div className="rounded-lg border border-dashed border-[#b9c2b8] bg-[#f6f5ef] p-6 text-center text-sm text-[#67746e]">
            {empty}
          </div>
        ) : (
          <div className="space-y-3">
            {items.slice(0, 8).map((item) => (
              <div key={item.id} className="rounded-lg border border-[#d7dbd2] bg-[#fbfaf5] p-4">
                <p className="font-medium text-[#14221d]">{item.title}</p>
                <p className="mt-1 text-sm text-[#67746e]">{item.meta}</p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
