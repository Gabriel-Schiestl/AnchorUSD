import { RiskDashboard } from "@/components/risk/risk-dashboard";

export default function RiskPage() {
  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">Risk Dashboard</h1>
        <p className="mt-2 text-muted-foreground">
          Risk and protocol health metrics
        </p>
      </div>
      <RiskDashboard />
    </div>
  );
}
