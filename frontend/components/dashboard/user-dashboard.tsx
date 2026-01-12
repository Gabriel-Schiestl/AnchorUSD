"use client";

import useSWR from "swr";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Progress } from "@/components/ui/progress";
import {
  Wallet,
  Shield,
  HeartPulse,
  TrendingUp,
  AlertTriangle,
} from "lucide-react";

// Fetcher para SWR
const fetcher = (url: string) => fetch(url).then((res) => res.json());

// Dados mockados para demonstração
const mockDashboardData = {
  balances: [
    { asset: "ETH", amount: "5.25", valueUsd: "12,500.00" },
    { asset: "WBTC", amount: "0.15", valueUsd: "6,300.00" },
    { asset: "USC", amount: "8,500.00", valueUsd: "8,500.00" },
  ],
  collateral: {
    total: "18,800.00",
    locked: "12,000.00",
    available: "6,800.00",
  },
  healthFactor: 1.85,
  debt: "6,500.00",
};

function getHealthFactorColor(hf: number) {
  if (hf >= 2) return "text-primary";
  if (hf >= 1.5) return "text-chart-3";
  if (hf >= 1.2) return "text-chart-5";
  return "text-destructive";
}

function getHealthFactorStatus(hf: number) {
  if (hf >= 2) return "Saudável";
  if (hf >= 1.5) return "Moderado";
  if (hf >= 1.2) return "Em Risco";
  return "Crítico";
}

export function UserDashboard() {
  const { data, isLoading } = useSWR("/api/dashboard", fetcher, {
    fallbackData: mockDashboardData,
  });

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  const healthPercent = Math.min((data.healthFactor / 3) * 100, 100);

  return (
    <div className="space-y-6">
      {/* Main metrics cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Collateral
            </CardTitle>
            <Shield className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-foreground">
              ${data.collateral.total}
            </div>
            <p className="text-xs text-muted-foreground">
              ${data.collateral.available} available
            </p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Debt
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-chart-5" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-foreground">
              ${data.debt}
            </div>
            <p className="text-xs text-muted-foreground">USC minted</p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Health Factor
            </CardTitle>
            <HeartPulse
              className={`h-4 w-4 ${getHealthFactorColor(data.healthFactor)}`}
            />
          </CardHeader>
          <CardContent>
            <div
              className={`text-2xl font-bold ${getHealthFactorColor(
                data.healthFactor
              )}`}
            >
              {data.healthFactor.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              {getHealthFactorStatus(data.healthFactor)}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Liquidation Limit
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-chart-5" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-foreground">$7,800.00</div>
            <p className="text-xs text-muted-foreground">Price: ETH $1,485</p>
          </CardContent>
        </Card>
      </div>

      {/*  Visual Health Factor */}
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <HeartPulse className="h-5 w-5 text-primary" />
            Health Factor
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Liquidation</span>
            <span className="text-sm text-muted-foreground">Healthy</span>
          </div>
          <Progress value={healthPercent} className="h-3" />
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>1.0</span>
            <span>1.5</span>
            <span>2.0</span>
            <span>2.5</span>
            <span>3.0+</span>
          </div>
        </CardContent>
      </Card>

      {/* Balance Table */}
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <Wallet className="h-5 w-5 text-primary" />
            Balance by Asset
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {data.balances.map(
              (balance: {
                asset: string;
                amount: string;
                valueUsd: string;
              }) => (
                <div
                  key={balance.asset}
                  className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 p-4"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                      <span className="font-mono text-lg text-primary">
                        {balance.asset === "ETH"
                          ? "Ξ"
                          : balance.asset === "WBTC"
                          ? "₿"
                          : "$"}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium text-foreground">
                        {balance.asset}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {balance.amount}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-mono font-medium text-foreground">
                      ${balance.valueUsd}
                    </p>
                  </div>
                </div>
              )
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Card key={i} className="border-border bg-card">
            <CardHeader className="pb-2">
              <Skeleton className="h-4 w-24" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-32" />
              <Skeleton className="mt-2 h-3 w-20" />
            </CardContent>
          </Card>
        ))}
      </div>
      <Card className="border-border bg-card">
        <CardHeader>
          <Skeleton className="h-6 w-40" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-3 w-full" />
        </CardContent>
      </Card>
    </div>
  );
}
