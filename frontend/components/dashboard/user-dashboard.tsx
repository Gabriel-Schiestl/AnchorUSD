"use client";

import useSWR from "swr";
import { useAccount } from "wagmi";
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
import { DashboardData, mockDashboardData } from "@/api/mocks/dashboard";
import { get } from "@/api/get";
import {
  getHealthFactorColor,
  getHealthFactorStatus,
  getHealthPercent,
} from "@/domain/healthFactor";
import { ConnectWalletPrompt } from "@/components/connect-wallet-prompt";

export function UserDashboard() {
  const { isConnected } = useAccount();
  const { data, isLoading } = useSWR("/dashboard", get<DashboardData>, {
    fallbackData: mockDashboardData,
  });

  if (!isConnected) {
    return (
      <ConnectWalletPrompt
        icon={AlertTriangle}
        description="Connect your wallet to view your dashboard"
      />
    );
  }

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="space-y-6">
      {/* Main metrics cards */}
      <div className="grid gap-4 md:grid-cols-3">
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
            <p className="text-xs text-muted-foreground">AUSD minted</p>
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
          <Progress
            value={getHealthPercent(data.healthFactor)}
            className="h-3"
          />
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
            Collateral Amount by Asset
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {data.collateralDeposited.map(
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
                        {balance.asset === "WETH"
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
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[...Array(3)].map((_, i) => (
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
