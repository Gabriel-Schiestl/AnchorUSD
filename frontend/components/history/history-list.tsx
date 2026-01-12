"use client";

import useSWR from "swr";
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

const fetcher = (url: string) => fetch(url).then((res) => res.json());

type TransactionType = "deposit" | "mint" | "burn" | "liquidation";

interface Transaction {
  id: string;
  type: TransactionType;
  amount: string;
  asset: string;
  timestamp: string;
  txHash: string;
  status: "completed" | "pending" | "failed";
}

const mockHistoryData = {
  deposits: [
    {
      id: "1",
      type: "deposit",
      amount: "2.5",
      asset: "ETH",
      timestamp: "2025-01-12T10:30:00Z",
      txHash: "0x1234...abcd",
      status: "completed",
    },
    {
      id: "2",
      type: "deposit",
      amount: "0.08",
      asset: "WBTC",
      timestamp: "2025-01-11T14:20:00Z",
      txHash: "0x5678...efgh",
      status: "completed",
    },
    {
      id: "3",
      type: "deposit",
      amount: "1.2",
      asset: "ETH",
      timestamp: "2025-01-10T09:15:00Z",
      txHash: "0x9abc...ijkl",
      status: "completed",
    },
  ],
  mintBurn: [
    {
      id: "4",
      type: "mint",
      amount: "5,000.00",
      asset: "USC",
      timestamp: "2025-01-12T11:00:00Z",
      txHash: "0xdef0...mnop",
      status: "completed",
    },
    {
      id: "5",
      type: "burn",
      amount: "1,500.00",
      asset: "USC",
      timestamp: "2025-01-11T16:45:00Z",
      txHash: "0x1111...qrst",
      status: "completed",
    },
    {
      id: "6",
      type: "mint",
      amount: "3,500.00",
      asset: "USC",
      timestamp: "2025-01-09T08:30:00Z",
      txHash: "0x2222...uvwx",
      status: "completed",
    },
  ],
  liquidations: [
    {
      id: "7",
      type: "liquidation",
      amount: "800.00",
      asset: "USC",
      timestamp: "2025-01-05T22:10:00Z",
      txHash: "0x3333...yzab",
      status: "completed",
    },
  ],
} as const;

const typeConfig = {
  deposit: {
    icon: PiggyBank,
    label: "Depósito",
    color: "bg-primary/10 text-primary",
  },
  mint: { icon: Coins, label: "Mint", color: "bg-chart-2/10 text-chart-2" },
  burn: {
    icon: Flame,
    label: "Burn",
    color: "bg-destructive/10 text-destructive",
  },
  liquidation: {
    icon: AlertTriangle,
    label: "Liquidação",
    color: "bg-chart-5/10 text-chart-5",
  },
};

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

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
                ? "Concluído"
                : tx.status === "pending"
                ? "Pendente"
                : "Falhou"}
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
            {tx.amount} {tx.asset}
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
  const { data, isLoading } = useSWR("/api/history", fetcher, {
    fallbackData: mockHistoryData,
  });

  if (isLoading) {
    return <HistorySkeleton />;
  }

  return (
    <Tabs defaultValue="deposits" className="w-full">
      <TabsList className="grid w-full grid-cols-3 bg-secondary">
        <TabsTrigger
          value="deposits"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
        >
          Deposits
        </TabsTrigger>
        <TabsTrigger
          value="mintburn"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
        >
          Mint / Burn
        </TabsTrigger>
        <TabsTrigger
          value="liquidations"
          className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
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
