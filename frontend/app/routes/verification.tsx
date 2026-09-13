import { CheckCircle } from "lucide-react";
import { Link } from "react-router";
import { BrandMark } from "~/components/brand-mark";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";

export default function VerificationPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#dbe3d9] px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex flex-col items-center">
          <BrandMark className="mb-4 scale-125" />
          <h1 className="font-heading text-4xl font-medium text-[#14221d]">Sabeel</h1>
        </div>
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center text-center">
              <CheckCircle className="mb-4 h-16 w-16 text-[#5f9d75]" />
              <h2 className="mb-2 text-xl font-semibold text-[#14221d]">Email Verification</h2>
              <p className="mb-6 text-[#67746e]">
                Email verification is managed securely by Zitadel.
              </p>
              <Link to="/dashboard">
                <Button>Go to Dashboard</Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
