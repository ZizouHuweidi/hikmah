import { ArrowRight } from "lucide-react";
import { BrandMark } from "~/components/brand-mark";
import { buttonVariants } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import { authURL } from "~/lib/auth";
import { cn } from "~/lib/utils";

export default function RegistrationPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#dbe3d9] px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex flex-col items-center">
          <BrandMark className="mb-4 scale-125" />
          <h1 className="font-heading text-4xl font-medium text-[#14221d]">Sabeel</h1>
          <p className="mt-2 text-[#67746e]">Create your library account</p>
        </div>
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Create Account</CardTitle>
            <CardDescription>
              Your account and sign-in methods are managed securely by Zitadel.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <a
              className={cn(buttonVariants({ size: "lg" }), "w-full")}
              href={authURL("/auth/register")}
            >
              Continue to registration
              <ArrowRight className="ml-2 h-4 w-4" />
            </a>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
