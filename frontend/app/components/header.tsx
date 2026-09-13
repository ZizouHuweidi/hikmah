import { Link } from "react-router";
import { Button } from "~/components/ui/button";
import { useAuthStore } from "~/lib/auth";

export default function Header() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);

  return (
    <header className="sticky top-0 z-50 border-b border-slate-200 bg-white/80 px-4 backdrop-blur-lg">
      <nav className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-3 gap-y-2 py-3 sm:py-4">
        <h2 className="m-0 flex-shrink-0 text-base font-semibold tracking-tight">
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-sm font-semibold text-slate-900 no-underline shadow-sm sm:px-4 sm:py-2"
          >
            <span className="h-2 w-2 rounded-full bg-gradient-to-r from-emerald-500 to-teal-400" />
            Sabeel
          </Link>
        </h2>

        <div className="ml-auto flex items-center gap-4">
          {isLoading ? null : isAuthenticated ? (
            <>
              <Link to="/dashboard">
                <Button variant="ghost" size="sm">
                  Dashboard
                </Button>
              </Link>
              <Link to="/settings">
                <Button variant="ghost" size="sm">
                  Settings
                </Button>
              </Link>
            </>
          ) : (
            <>
              <Link to="/login">
                <Button variant="ghost" size="sm">
                  Sign In
                </Button>
              </Link>
              <Link to="/registration">
                <Button size="sm" className="bg-emerald-600 hover:bg-emerald-700">
                  Get Started
                </Button>
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
}
