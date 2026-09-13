import { Link } from "react-router";
import { BrandMark } from "~/components/brand-mark";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";

export default function RecoveryPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#dbe3d9] px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex flex-col items-center">
          <BrandMark className="mb-4 scale-125" />
          <h1 className="font-heading text-4xl font-medium text-[#14221d]">Sabeel</h1>
        </div>
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Account Recovery</CardTitle>
            <CardDescription>
              Continue to sign in and choose the recovery option provided by Zitadel.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link to="/login">
              <Button className="w-full">Back to Sign In</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
