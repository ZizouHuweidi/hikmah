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
      <div className="flex min-h-screen items-center justify-center bg-[#f6f5ef]">
        <div className="animate-pulse font-medium text-[#1d4236]">Loading...</div>
      </div>
    );
  }

  const features = [
    {
      icon: <Library className="h-8 w-8 text-[#286f62]" />,
      title: "Organize Your Knowledge",
      description: "Track books, papers, podcasts, videos, and articles in one unified library.",
    },
    {
      icon: <BookOpen className="h-8 w-8 text-[#286f62]" />,
      title: "Smart Annotations",
      description: "Create structured notes and annotations tied to source materials.",
    },
    {
      icon: <Brain className="h-8 w-8 text-[#286f62]" />,
      title: "Connect Ideas",
      description: "Use tags, topics, and taxonomies to connect your sources and insights.",
    },
    {
      icon: <Users className="h-8 w-8 text-[#286f62]" />,
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
    <main className="min-h-screen bg-[#f6f5ef]">
      <section className="border-b border-[#d7dbd2] bg-[#dbe3d9] px-6 py-20">
        <div className="mx-auto max-w-6xl text-center">
          <div className="mb-8 inline-flex items-center gap-2 border border-[#b9c2b8] bg-[#f6f5ef] px-4 py-2 text-sm font-medium text-[#1d4236]">
            <Sparkles className="h-4 w-4" />
            <span>A path for organized learning</span>
          </div>

          <h1 className="mb-6 font-heading text-5xl font-medium text-[#14221d] md:text-7xl">
            <span className="block">Sabeel</span>
          </h1>
          <p className="mx-auto mb-4 max-w-2xl text-xl text-[#67746e]">
            A modern platform for organizing, engaging with, and tracking knowledge sources across
            media.
          </p>
          <p className="mx-auto mb-10 max-w-xl text-[#67746e]">
            Built for readers who want a clearer path through books, notes, reviews, and ideas.
          </p>

          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            {isAuthenticated ? (
              <Link to="/dashboard">
                <Button size="lg" className="px-8 py-6 text-lg font-semibold">
                  Go to Dashboard
                </Button>
              </Link>
            ) : (
              <>
                <Link to="/registration">
                  <Button size="lg" className="px-8 py-6 text-lg font-semibold">
                    Start Your Journey
                  </Button>
                </Link>
                <Link to="/login">
                  <Button
                    size="lg"
                    variant="outline"
                    className="border-[#b9c2b8] px-8 py-6 text-lg font-semibold hover:border-[#1d4236] hover:text-[#1d4236]"
                  >
                    Sign In
                  </Button>
                </Link>
              </>
            )}
          </div>

          {isAuthenticated && user.email && (
            <p className="mt-4 font-medium text-[#286f62]">
              Welcome back, {user.firstName || user.email}!
            </p>
          )}
        </div>
      </section>

      <section className="border-b border-[#d7dbd2] bg-[#f6f5ef] px-6 py-12">
        <div className="mx-auto max-w-6xl">
          <p className="mb-8 text-center text-sm font-semibold uppercase tracking-wider text-[#67746e]">
            Track Everything You Learn From
          </p>
          <div className="flex flex-wrap items-center justify-center gap-8">
            {supportedTypes.map((type) => (
              <div
                key={type.label}
                className="flex items-center gap-3 border-l-2 border-[#c8e05c] bg-[#fbfaf5] px-6 py-3"
              >
                <span className="text-[#286f62]">{type.icon}</span>
                <span className="font-medium text-[#14221d]">{type.label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="bg-[#f6f5ef] px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <div className="mb-16 text-center">
            <h2 className="mb-4 text-3xl font-bold text-[#14221d]">Your Knowledge, Organized</h2>
            <p className="mx-auto max-w-2xl text-lg text-[#67746e]">
              Build a personal library that grows with you.
            </p>
          </div>
          <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="rounded-sm border border-[#d7dbd2] bg-[#fbfaf5] p-8 transition-colors duration-200 hover:border-[#1d4236]"
              >
                <div className="mb-6 w-fit border border-[#d7dbd2] bg-[#e9ebe4] p-3">
                  {feature.icon}
                </div>
                <h3 className="mb-3 text-xl font-semibold text-[#14221d]">{feature.title}</h3>
                <p className="leading-relaxed text-[#67746e]">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
