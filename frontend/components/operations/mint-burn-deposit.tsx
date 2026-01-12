"use client";

import { useState } from "react";
import { useAccount } from "wagmi";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { ArrowDownUp, Coins, PiggyBank, Flame, Loader2 } from "lucide-react";

const collateralAssets = [
  { symbol: "ETH", name: "Ethereum", icon: "Ξ" },
  { symbol: "WBTC", name: "Wrapped Bitcoin", icon: "₿" },
];

export function MintBurnDeposit() {
  const { isConnected } = useAccount();
  const [activeTab, setActiveTab] = useState("deposit");
  const [selectedAsset, setSelectedAsset] = useState("ETH");
  const [amount, setAmount] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (action: string) => {
    setIsLoading(true);
    // Simular transação
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setIsLoading(false);
    setAmount("");
  };

  if (!isConnected) {
    return (
      <Card className="border-border bg-card">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <Coins className="h-8 w-8 text-primary" />
          </div>
          <h3 className="mb-2 text-xl font-semibold text-foreground">
            Connect your wallet
          </h3>
          <p className="mb-6 text-center text-muted-foreground">
            To perform mint, burn, or deposit operations, you need to connect
            your wallet.
          </p>
          <ConnectButton />
        </CardContent>
      </Card>
    );
  }

  return (
    <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
      <TabsList className="grid w-full grid-cols-3 bg-secondary">
        <TabsTrigger
          value="deposit"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
        >
          <PiggyBank className="h-4 w-4" />
          Deposit
        </TabsTrigger>
        <TabsTrigger
          value="mint"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
        >
          <Coins className="h-4 w-4" />
          Mint
        </TabsTrigger>
        <TabsTrigger
          value="burn"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
        >
          <Flame className="h-4 w-4" />
          Burn
        </TabsTrigger>
      </TabsList>

      <TabsContent value="deposit">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <PiggyBank className="h-5 w-5 text-primary" />
              Deposit Collateral
            </CardTitle>
            <CardDescription>
              Deposit your assets as collateral to mint stablecoins
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="collateral-asset">Collateral Asset</Label>
              <Select value={selectedAsset} onValueChange={setSelectedAsset}>
                <SelectTrigger id="collateral-asset" className="bg-secondary">
                  <SelectValue placeholder="Select an asset" />
                </SelectTrigger>
                <SelectContent>
                  {collateralAssets.map((asset) => (
                    <SelectItem key={asset.symbol} value={asset.symbol}>
                      <div className="flex items-center gap-2">
                        <span className="font-mono">{asset.icon}</span>
                        <span>{asset.symbol}</span>
                        <span className="text-muted-foreground">
                          - {asset.name}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="deposit-amount">Amount</Label>
              <div className="relative">
                <Input
                  id="deposit-amount"
                  type="number"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="bg-secondary pr-16"
                />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  {selectedAsset}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                Available balance: 10.5 {selectedAsset}
              </p>
            </div>
            <Button
              className="w-full"
              size="lg"
              onClick={() => handleSubmit("deposit")}
              disabled={!amount || isLoading}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <PiggyBank className="mr-2 h-4 w-4" />
                  Deposit {selectedAsset}
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="mint">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <Coins className="h-5 w-5 text-primary" />
              Mint Stablecoin
            </CardTitle>
            <CardDescription>
              Mint stablecoins using your deposited collateral
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border border-border bg-secondary/50 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Available Collateral Value
                </span>
                <span className="font-mono text-foreground">$15,420.00</span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Maximum Mintable
                </span>
                <span className="font-mono text-primary">$10,280.00</span>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="mint-amount">Amount to Mint</Label>
              <div className="relative">
                <Input
                  id="mint-amount"
                  type="number"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="bg-secondary pr-16"
                />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  USC
                </span>
              </div>
            </div>
            <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
              <div className="flex items-center gap-2 text-sm">
                <ArrowDownUp className="h-4 w-4 text-primary" />
                <span className="text-muted-foreground">
                  Health Factor after mint:
                </span>
                <span className="font-mono font-semibold text-primary">
                  1.85
                </span>
              </div>
            </div>
            <Button
              className="w-full"
              size="lg"
              onClick={() => handleSubmit("mint")}
              disabled={!amount || isLoading}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <Coins className="mr-2 h-4 w-4" />
                  Mint USC
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="burn">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-foreground">
              <Flame className="h-5 w-5 text-destructive" />
              Burn Stablecoin
            </CardTitle>
            <CardDescription>
              Burn stablecoins to release your collateral
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border border-border bg-secondary/50 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  USC Balance
                </span>
                <span className="font-mono text-foreground">5,000.00 USC</span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Total Debt
                </span>
                <span className="font-mono text-destructive">3,500.00 USC</span>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="burn-amount">Amount to Burn</Label>
              <div className="relative">
                <Input
                  id="burn-amount"
                  type="number"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="bg-secondary pr-16"
                />
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                  USC
                </span>
              </div>
            </div>
            <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
              <div className="flex items-center gap-2 text-sm">
                <ArrowDownUp className="h-4 w-4 text-primary" />
                <span className="text-muted-foreground">
                  Health Factor after burn:
                </span>
                <span className="font-mono font-semibold text-primary">
                  2.45
                </span>
              </div>
            </div>
            <Button
              variant="destructive"
              className="w-full"
              size="lg"
              onClick={() => handleSubmit("burn")}
              disabled={!amount || isLoading}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <Flame className="mr-2 h-4 w-4" />
                  Burn USC
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  );
}
