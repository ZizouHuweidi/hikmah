import { Link } from "react-router";
import { BrandMark } from "~/components/brand-mark";
import { Button } from "~/components/ui/button";
import { useAuthStore } from "~/lib/auth";

export default function Header() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);

  return (
    <header className="sticky top-0 z-50 border-b border-[#d7dbd2] bg-[#f6f5ef]/95 px-4 backdrop-blur-lg">
      <nav className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-3 gap-y-2 py-3 sm:py-4">
        <h2 className="m-0 flex-shrink-0 text-base font-semibold tracking-tight">
          <Link to="/" className="inline-flex items-center gap-3 text-[#14221d] no-underline">
            <BrandMark className="scale-75" />
            <span className="font-heading text-xl font-medium">Sabeel</span>
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
                <Button size="sm">Get Started</Button>
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
}
