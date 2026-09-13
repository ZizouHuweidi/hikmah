import { ArrowRight } from "lucide-react";
import { BrandMark } from "~/components/brand-mark";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import { useAuthStore } from "~/lib/auth";

export default function LoginPage() {
  const login = useAuthStore((state) => state.login);

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#dbe3d9] px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex flex-col items-center">
          <BrandMark className="mb-4 scale-125" />
          <h1 className="font-heading text-4xl font-medium text-[#14221d]">Sabeel</h1>
          <p className="mt-2 text-[#67746e]">Welcome back to your library</p>
        </div>
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Sign In</CardTitle>
            <CardDescription>Continue securely with your Sabeel identity.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button className="w-full" size="lg" onClick={() => login()}>
              Continue to sign in
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
