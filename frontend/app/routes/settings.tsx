import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Library, Loader2, Mail, Shield, User } from "lucide-react";
import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { getProfile, updateProfile } from "~/lib/api";
import { useAuthStore } from "~/lib/auth";

export default function SettingsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { isAuthenticated, isLoading, user } = useAuthStore();
  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [publicProfile, setPublicProfile] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) navigate("/login");
  }, [isAuthenticated, isLoading, navigate]);

  const profileQuery = useQuery({
    queryKey: ["profile"],
    enabled: isAuthenticated,
    queryFn: getProfile,
  });

  useEffect(() => {
    if (!profileQuery.data) return;
    setDisplayName(profileQuery.data.display_name || "");
    setBio(profileQuery.data.bio || "");
    setPublicProfile(profileQuery.data.public_profile);
  }, [profileQuery.data]);

  const updateMutation = useMutation({
    mutationFn: () =>
      updateProfile({
        display_name: displayName.trim() || undefined,
        bio: bio.trim() || undefined,
        public_profile: publicProfile,
      }),
    onSuccess: async () => {
      setMessage("Profile saved");
      await queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    onError: (err) => setError(err instanceof Error ? err.message : "Failed to save profile"),
  });

  const handleSave = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setMessage(null);
    updateMutation.mutate();
  };

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#f6f5ef]">
        <Loader2 className="h-8 w-8 animate-spin text-[#286f62]" />
      </div>
    );
  }

  const profile = profileQuery.data;
  const loadingProfile = profileQuery.isLoading;

  return (
    <div className="min-h-screen bg-[#f6f5ef] py-8">
      <div className="mx-auto max-w-2xl px-4">
        <div className="mb-8 flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="rounded-xl bg-[#1d4236] p-2">
              <Library className="h-6 w-6 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-[#14221d]">Account Settings</h1>
          </div>
          <Link to="/dashboard">
            <Button variant="outline">Dashboard</Button>
          </Link>
        </div>

        {(error || profileQuery.error) && (
          <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error ||
              (profileQuery.error instanceof Error
                ? profileQuery.error.message
                : "Failed to load profile")}
          </div>
        )}
        {message && (
          <div className="mb-6 rounded-lg border border-[#c8e05c] bg-[#e4ecc9] px-4 py-3 text-sm text-[#1d4236]">
            {message}
          </div>
        )}

        <Card>
          <CardContent className="space-y-6 pt-6">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#e4ecc9]">
                <User className="h-6 w-6 text-[#286f62]" />
              </div>
              <div>
                <p className="font-medium text-[#14221d]">{user.username}</p>
                <p className="text-sm text-[#67746e]">Username</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#e4ecc9]">
                <Mail className="h-6 w-6 text-[#286f62]" />
              </div>
              <div>
                <p className="font-medium text-[#14221d]">{user.email || "No email"}</p>
                <p className="text-sm text-[#67746e]">Email address</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#e4ecc9]">
                <Shield className="h-6 w-6 text-[#286f62]" />
              </div>
              <div>
                <p className="font-medium text-[#14221d]">Protected by Zitadel</p>
                <p className="text-sm text-[#67746e]">
                  Manage sign-in methods with your identity account.
                </p>
              </div>
            </div>

            <form className="space-y-4 border-t border-[#d7dbd2] pt-6" onSubmit={handleSave}>
              <div>
                <label className="mb-2 block text-sm font-medium text-[#14221d]">
                  Display name
                </label>
                <Input
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                  placeholder="How your public profile should appear"
                  disabled={loadingProfile}
                />
              </div>
              <div>
                <label className="mb-2 block text-sm font-medium text-[#14221d]">Bio</label>
                <textarea
                  className="min-h-28 w-full rounded-md border border-[#b9c2b8] bg-[#fbfaf5] px-3 py-2 text-sm"
                  value={bio}
                  onChange={(event) => setBio(event.target.value)}
                  placeholder="What are you reading, researching, or collecting?"
                  disabled={loadingProfile}
                />
              </div>
              <label className="flex items-start gap-3 rounded-lg border border-[#d7dbd2] bg-[#f6f5ef] p-4 text-sm text-[#14221d]">
                <input
                  type="checkbox"
                  className="mt-1"
                  checked={publicProfile}
                  onChange={(event) => setPublicProfile(event.target.checked)}
                  disabled={loadingProfile}
                />
                <span>
                  <span className="block font-medium text-[#14221d]">Make my profile public</span>
                  Public profiles can anchor public notes, reviews, collections, and library items.
                </span>
              </label>
              {profile?.public_profile && user.username && (
                <p className="text-sm text-[#67746e]">
                  Public profile:{" "}
                  <Link
                    to={`/users/${user.username}/profile`}
                    className="font-medium text-[#1d4236] underline-offset-4 hover:underline"
                  >
                    /users/{user.username}/profile
                  </Link>
                </p>
              )}
              <Button type="submit" disabled={updateMutation.isPending || loadingProfile}>
                {updateMutation.isPending ? "Saving..." : "Save profile"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
