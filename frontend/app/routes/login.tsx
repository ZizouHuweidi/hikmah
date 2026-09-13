import { ArrowRight, Library } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { useAuthStore } from "~/lib/auth";

export default function LoginPage() {
  const navigate = useNavigate();
  const login = useAuthStore((state) => state.login);
  const [loginValue, setLoginValue] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);
    try {
      await login(loginValue, password);
      navigate("/dashboard");
    } catch {
      setError("Invalid email/username or password.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-slate-50 to-slate-100 px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex flex-col items-center">
          <div className="mb-4 rounded-xl bg-emerald-600 p-3">
            <Library className="h-8 w-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-slate-900">Sabeel</h1>
          <p className="mt-2 text-slate-600">Welcome back to your library</p>
        </div>
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Sign In</CardTitle>
            <CardDescription>Sign in with your email or username.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <Label htmlFor="login">Email or username</Label>
                <Input
                  id="login"
                  value={loginValue}
                  onChange={(event) => setLoginValue(event.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                />
              </div>
              {error && <p className="text-sm text-red-600">{error}</p>}
              <Button className="w-full" size="lg" disabled={isSubmitting}>
                Sign In
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              <p className="text-center text-sm text-slate-500">
                <Link to="/recovery" className="text-emerald-700 hover:underline">
                  Forgot password?
                </Link>
              </p>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
