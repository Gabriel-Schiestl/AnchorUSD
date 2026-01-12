import { UserDashboard } from "@/components/dashboard/user-dashboard";

export default function DashboardPage() {
  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
        <p className="mt-2 text-muted-foreground">
          Overview of your assets and collateral position
        </p>
      </div>
      <UserDashboard />
    </div>
  );
}
