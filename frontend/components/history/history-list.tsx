"use client";

import useSWR from "swr";
import { useAccount } from "wagmi";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  PiggyBank,
  Coins,
  Flame,
  AlertTriangle,
  ExternalLink,
} from "lucide-react";
import { get, HistoryData } from "@/api/get";
import Transaction from "@/models/Transaction";
import { mockHistoryData } from "@/api/mocks/history";
import { formatDate } from "@/lib/date";
import { typeConfig } from "@/models/TypeConfig";
import { ConnectWalletPrompt } from "@/components/connect-wallet-prompt";
import { formatFromWeiPrecise } from "@/lib/utils";

function TransactionItem({ tx }: { tx: Transaction }) {
  const config = typeConfig[tx.type];
  const Icon = config.icon;

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 p-4 transition-colors hover:bg-secondary/50">
      <div className="flex items-center gap-4">
        <div
          className={`flex h-10 w-10 items-center justify-center rounded-full ${config.color}`}
        >
          <Icon className="h-5 w-5" />
        </div>
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium text-foreground">{config.label}</span>
            <Badge variant="outline" className="text-xs">
              {tx.status === "completed"
                ? "Completed"
                : tx.status === "pending"
                  ? "Pending"
                  : "Failed"}
            </Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            {formatDate(tx.timestamp)}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="text-right">
          <p className="font-mono font-medium text-foreground">
            {tx.type === "burn" ? "-" : "+"}
            {formatFromWeiPrecise(tx.amount, 18, 6)} {tx.asset}
          </p>
          <a
            href={`https://etherscan.io/tx/${tx.txHash}`}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1 text-xs text-primary hover:underline"
          >
            {tx.txHash}
            <ExternalLink className="h-3 w-3" />
          </a>
        </div>
      </div>
    </div>
  );
}

export function HistoryList() {
  const { isConnected, address } = useAccount();
  const { data, isLoading } = useSWR(
    address && isConnected ? `/api/history/${address}` : null,
    () => get<HistoryData>(`/api/history/${address}`),
    {
      fallbackData: mockHistoryData,
    },
  );

  if (!isConnected) {
    return (
      <ConnectWalletPrompt
        icon={AlertTriangle}
        description="Connect your wallet to view your transaction history"
      />
    );
  }

  if (isLoading) {
    return <HistorySkeleton />;
  }

  return (
    <Tabs defaultValue="deposits" className="w-full">
      <TabsList className="grid w-full grid-cols-3 bg-secondary">
        <TabsTrigger
          value="deposits"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          Deposits
        </TabsTrigger>
        <TabsTrigger
          value="mintburn"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          Mint / Burn
        </TabsTrigger>
        <TabsTrigger
          value="liquidations"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          Liquidations
        </TabsTrigger>
      </TabsList>

      <TabsContent value="deposits" className="mt-6">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <PiggyBank className="h-5 w-5 text-primary" />
              Deposit History
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {data.deposits.length > 0 ? (
              data.deposits.map((tx: Transaction) => (
                <TransactionItem key={tx.id} tx={tx} />
              ))
            ) : (
              <p className="py-8 text-center text-muted-foreground">
                No deposits found
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="mintburn" className="mt-6">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <Coins className="h-5 w-5 text-primary" />
              Mint / Burn history
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {data.mintBurn.length > 0 ? (
              data.mintBurn.map((tx: Transaction) => (
                <TransactionItem key={tx.id} tx={tx} />
              ))
            ) : (
              <p className="py-8 text-center text-muted-foreground">
                No operations found
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="liquidations" className="mt-6">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <AlertTriangle className="h-5 w-5 text-chart-5" />
              Liquidations History
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {data.liquidations.length > 0 ? (
              data.liquidations.map((tx: Transaction) => (
                <TransactionItem key={tx.id} tx={tx} />
              ))
            ) : (
              <p className="py-8 text-center text-muted-foreground">
                No liquidations found
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  );
}

function HistorySkeleton() {
  return (
    <Card className="border-border bg-card">
      <CardHeader>
        <Skeleton className="h-6 w-48" />
      </CardHeader>
      <CardContent className="space-y-3">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </CardContent>
    </Card>
  );
}
