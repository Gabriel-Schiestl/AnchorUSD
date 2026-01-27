"use client";

import useSWR from "swr";
import { useAccount } from "wagmi";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Progress } from "@/components/ui/progress";
import {
  AlertTriangle,
  Shield,
  Coins,
  Users,
  TrendingDown,
  Activity,
} from "lucide-react";
import { ConnectWalletPrompt } from "@/components/connect-wallet-prompt";
import { mockRiskData } from "@/api/mocks/user";
import { get, LiquidatableUser, MetricsData } from "@/api/get";
import { getHealthFactorColor } from "@/domain/healthFactor";
import { formatFromWei, formatFromWeiPrecise } from "@/lib/utils";

export function RiskDashboard() {
  const { isConnected } = useAccount();
  const { data, isLoading } = useSWR(
    isConnected ? `/api/metrics/dashboard` : null,
    () => get<MetricsData>(`/api/metrics/dashboard`),
    {
      fallbackData: mockRiskData,
    },
  );

  if (!isConnected) {
    return (
      <ConnectWalletPrompt
        icon={AlertTriangle}
        description="Connect your wallet to view the risk dashboard"
      />
    );
  }

  if (isLoading) {
    return <RiskSkeleton />;
  }

  const riskPercentage =
    (data.protocolHealth.usersAtRisk / data.protocolHealth.totalUsers) * 100;

  return (
    <div className="space-y-6">
      {/* Main metrics */}
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
              ${formatFromWeiPrecise(data.totalCollateral.value, 8, 2)}
            </div>
            <p className="text-xs text-muted-foreground">
              Locked in the protocol
            </p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Stablecoin Supply
            </CardTitle>
            <Coins className="h-4 w-4 text-chart-2" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-foreground">
              ${parseFloat(formatFromWei(data.stableSupply.total)).toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              {data.stableSupply.backing}% collateralized
            </p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Users at Risk
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-chart-5" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-chart-5">
              {data.protocolHealth.usersAtRisk}
            </div>
            <p className="text-xs text-muted-foreground">
              of {data.protocolHealth.totalUsers.toLocaleString()} users
            </p>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Average Health Factor
            </CardTitle>
            <Activity className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-primary">
              {data.protocolHealth.averageHealthFactor.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">Protocol health</p>
          </CardContent>
        </Card>
      </div>

      {/* Breakdown of collateral */}
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <Shield className="h-5 w-5 text-primary" />
            Breakdown of Collateral
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {data.totalCollateral.breakdown.map(
            (item: {
              asset: string;
              amount: string;
              valueUsd: string;
              percentage: number;
            }) => (
              <div key={item.asset} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-foreground">
                      {item.asset}
                    </span>
                    <Badge variant="outline" className="text-xs">
                      {item.amount}
                    </Badge>
                  </div>
                  <div className="text-right">
                    <span className="font-mono text-foreground">
                      ${formatFromWeiPrecise(item.valueUsd, 8, 2)}
                    </span>
                    <span className="ml-2 text-sm text-muted-foreground">
                      ({item.percentage}%)
                    </span>
                  </div>
                </div>
                <Progress value={item.percentage} className="h-2" />
              </div>
            ),
          )}
        </CardContent>
      </Card>

      {/* Liquidatable Users */}
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <TrendingDown className="h-5 w-5 text-destructive" />
            Liquidatable Users
            <Badge variant="destructive" className="ml-2">
              {data.liquidatableUsers.length} positions
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {data.liquidatableUsers.length > 0 ? (
            <div className="space-y-3">
              {data.liquidatableUsers.map((user: LiquidatableUser) => (
                <div
                  key={user.address}
                  className="flex items-center justify-between rounded-lg border border-destructive/20 bg-destructive/5 p-4"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
                      <Users className="h-5 w-5 text-destructive" />
                    </div>
                    <div>
                      <p className="font-mono text-sm text-foreground">
                        {user.address}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        Collateral: ${user.collateralUsd} | Debt: $
                        {user.debtUsd}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p
                      className={`font-mono font-bold ${getHealthFactorColor(
                        parseFloat(user.healthFactor),
                      )}`}
                    >
                      HF: {parseFloat(user.healthFactor).toFixed(2)}
                    </p>
                    <p className="text-xs text-destructive">
                      Liquidatable: ${user.liquidationAmount}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12">
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
                <Shield className="h-8 w-8 text-primary" />
              </div>
              <p className="text-lg font-medium text-foreground">
                No liquidatable users
              </p>
              <p className="text-sm text-muted-foreground">
                All positions are healthy
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Protocol Risk Metrics */}
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <Activity className="h-5 w-5 text-primary" />
            Protocol Health
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-lg border border-border bg-secondary/30 p-4">
              <p className="text-sm text-muted-foreground">
                Collateralization Ratio
              </p>
              <p className="mt-1 text-2xl font-bold text-primary">
                {data.protocolHealth.collateralizationRatio}%
              </p>
            </div>
            <div className="rounded-lg border border-border bg-secondary/30 p-4">
              <p className="text-sm text-muted-foreground">Users at Risk (%)</p>
              <p className="mt-1 text-2xl font-bold text-chart-5">
                {riskPercentage.toFixed(2)}%
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function RiskSkeleton() {
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
          <Skeleton className="h-6 w-48" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    </div>
  );
}
