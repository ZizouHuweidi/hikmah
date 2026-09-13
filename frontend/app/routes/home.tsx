import { Book, BookOpen, Brain, Library, Quote, Scroll, Sparkles, Users } from "lucide-react";
import { Link } from "react-router";
import { Button } from "~/components/ui/button";
import { useAuthStore } from "~/lib/auth";
export function meta() {
  return [
    { title: "Sabeel - Knowledge Library" },
    { name: "description", content: "Organize, track, and share your knowledge library." },
  ];
}

export default function Home() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const user = useAuthStore((state) => state.user);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-slate-50 to-slate-100">
        <div className="animate-pulse font-medium text-emerald-600">Loading...</div>
      </div>
    );
  }

  const features = [
    {
      icon: <Library className="h-8 w-8 text-emerald-600" />,
      title: "Organize Your Knowledge",
      description: "Track books, papers, podcasts, videos, and articles in one unified library.",
    },
    {
      icon: <BookOpen className="h-8 w-8 text-emerald-600" />,
      title: "Smart Annotations",
      description: "Create structured notes and annotations tied to source materials.",
    },
    {
      icon: <Brain className="h-8 w-8 text-emerald-600" />,
      title: "Connect Ideas",
      description: "Use tags, topics, and taxonomies to connect your sources and insights.",
    },
    {
      icon: <Users className="h-8 w-8 text-emerald-600" />,
      title: "Build Your Identity",
      description: "Create a public or private knowledge profile with reviews and curated lists.",
    },
  ];

  const supportedTypes = [
    { icon: <Book className="h-6 w-6" />, label: "Books" },
    { icon: <Scroll className="h-6 w-6" />, label: "Academic Papers" },
    { icon: <Quote className="h-6 w-6" />, label: "Articles" },
    { icon: <Sparkles className="h-6 w-6" />, label: "Podcasts & Videos" },
  ];

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-50 to-slate-100">
      <section className="relative overflow-hidden px-6 py-24">
        <div className="absolute inset-0 bg-gradient-to-br from-emerald-50 via-teal-50 to-cyan-50 opacity-70" />
        <div className="absolute -left-20 -top-24 h-96 w-96 animate-pulse rounded-full bg-emerald-200 opacity-30 mix-blend-multiply blur-3xl" />
        <div className="absolute -bottom-20 -right-20 h-96 w-96 animate-pulse rounded-full bg-teal-200 opacity-30 mix-blend-multiply blur-3xl" />

        <div className="relative mx-auto max-w-6xl text-center">
          <div className="mb-8 inline-flex items-center gap-2 rounded-full bg-emerald-100 px-4 py-2 text-sm font-medium text-emerald-800">
            <Sparkles className="h-4 w-4" />
            <span>A path for organized learning</span>
          </div>

          <h1 className="mb-6 text-5xl font-bold tracking-tight text-slate-900 md:text-7xl">
            <span className="block bg-gradient-to-r from-emerald-600 to-teal-600 bg-clip-text text-transparent">
              Sabeel
            </span>
          </h1>
          <p className="mx-auto mb-4 max-w-2xl text-xl text-slate-600">
            A modern platform for organizing, engaging with, and tracking knowledge sources across
            media.
          </p>
          <p className="mx-auto mb-10 max-w-xl text-slate-500">
            Built for readers who want a clearer path through books, notes, reviews, and ideas.
          </p>

          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            {isAuthenticated ? (
              <Link to="/dashboard">
                <Button
                  size="lg"
                  className="rounded-full bg-emerald-600 px-8 py-6 text-lg font-semibold text-white shadow-lg hover:bg-emerald-700"
                >
                  Go to Dashboard
                </Button>
              </Link>
            ) : (
              <>
                <Link to="/registration">
                  <Button
                    size="lg"
                    className="rounded-full bg-emerald-600 px-8 py-6 text-lg font-semibold text-white shadow-lg hover:bg-emerald-700"
                  >
                    Start Your Journey
                  </Button>
                </Link>
                <Link to="/login">
                  <Button
                    size="lg"
                    variant="outline"
                    className="rounded-full border-2 border-slate-300 px-8 py-6 text-lg font-semibold hover:border-emerald-500 hover:text-emerald-700"
                  >
                    Sign In
                  </Button>
                </Link>
              </>
            )}
          </div>

          {isAuthenticated && user.email && (
            <p className="mt-4 font-medium text-emerald-600">
              Welcome back, {user.firstName || user.email}!
            </p>
          )}
        </div>
      </section>

      <section className="border-y border-slate-200 bg-white/50 px-6 py-12">
        <div className="mx-auto max-w-6xl">
          <p className="mb-8 text-center text-sm font-semibold uppercase tracking-wider text-slate-500">
            Track Everything You Learn From
          </p>
          <div className="flex flex-wrap items-center justify-center gap-8">
            {supportedTypes.map((type) => (
              <div
                key={type.label}
                className="flex items-center gap-3 rounded-full bg-white px-6 py-3 shadow-sm"
              >
                <span className="text-emerald-600">{type.icon}</span>
                <span className="font-medium text-slate-700">{type.label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="px-6 py-24">
        <div className="mx-auto max-w-6xl">
          <div className="mb-16 text-center">
            <h2 className="mb-4 text-3xl font-bold text-slate-900">Your Knowledge, Organized</h2>
            <p className="mx-auto max-w-2xl text-lg text-slate-600">
              Build a personal library that grows with you.
            </p>
          </div>
          <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="rounded-2xl border border-slate-200 bg-white p-8 transition-all duration-300 hover:border-emerald-200 hover:shadow-lg"
              >
                <div className="mb-6 w-fit rounded-xl bg-slate-50 p-3">{feature.icon}</div>
                <h3 className="mb-3 text-xl font-semibold text-slate-900">{feature.title}</h3>
                <p className="leading-relaxed text-slate-600">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
