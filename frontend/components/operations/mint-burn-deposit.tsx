"use client";

import { useState, useEffect } from "react";
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
import { ConnectWalletPrompt } from "@/components/connect-wallet-prompt";
import {
  ArrowDownUp,
  Coins,
  PiggyBank,
  Flame,
  Loader2,
  RefreshCw,
  AlertTriangle,
} from "lucide-react";
import { useWalletBalance } from "@/hooks/useWalletBalance";
import { useAUSDEngine } from "@/hooks/useAUSDEngine";
import { collateralAssets, TOKEN_ADDRESSES } from "@/lib/constants";
import { formatFromWei, formatFromWeiPrecise } from "@/lib/utils";

export function MintBurnDeposit() {
  const { isConnected } = useAccount();
  const [activeTab, setActiveTab] = useState("deposit");
  const [selectedAsset, setSelectedAsset] = useState("WETH");
  const [amount, setAmount] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [healthFactorProjection, setHealthFactorProjection] = useState<
    string | null
  >(null);

  const selectedTokenAddress = collateralAssets.find(
    (asset) => asset.symbol === selectedAsset
  )?.address;

  const {
    balance: assetBalance,
    isLoading: isLoadingBalance,
    refresh: refreshAssetBalance,
  } = useWalletBalance(selectedTokenAddress);

  const { balance: ausdBalance, refresh: refreshAusdBalance } =
    useWalletBalance(TOKEN_ADDRESSES.AUSD);

  const {
    engineData,
    isLoading: isLoadingEngine,
    refresh: refreshEngineData,
    calculateHealthFactorAfterMint,
    calculateHealthFactorAfterBurn,
    calculateHealthFactorAfterDeposit,
  } = useAUSDEngine();

  useEffect(() => {
    const updateHealthFactorProjection = async () => {
      if (!amount || isNaN(parseFloat(amount))) {
        setHealthFactorProjection(null);
        return;
      }

      try {
        let projection = null;

        if (activeTab === "mint") {
          projection = await calculateHealthFactorAfterMint(amount);
        } else if (activeTab === "burn") {
          projection = await calculateHealthFactorAfterBurn(amount);
        } else if (activeTab === "deposit" && selectedTokenAddress) {
          projection = await calculateHealthFactorAfterDeposit(
            selectedTokenAddress,
            amount
          );
        }

        if (projection) {
          setHealthFactorProjection(projection.healthFactorAfter);
        }
      } catch (error) {
        console.error("Error calculating health factor:", error);
        setHealthFactorProjection(null);
      }
    };

    const debounce = setTimeout(updateHealthFactorProjection, 500);
    return () => clearTimeout(debounce);
  }, [amount, activeTab, selectedTokenAddress]);

  const handleSubmit = async (action: string) => {
    setIsLoading(true);
    try {
      // Implement logic to interact with smart contracts here
      // If deposit, call approve function for the selected asset first

      await new Promise((resolve) => setTimeout(resolve, 2000));

      await Promise.all([
        refreshAssetBalance(),
        refreshAusdBalance(),
        refreshEngineData(),
      ]);

      setAmount("");
      setHealthFactorProjection(null);
    } catch (error) {
      console.error(`Error during ${action}:`, error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRefresh = async () => {
    await Promise.all([
      refreshAssetBalance(),
      refreshAusdBalance(),
      refreshEngineData(),
    ]);
  };

  if (!isConnected) {
    return (
      <ConnectWalletPrompt
        icon={AlertTriangle}
        description="To perform mint, burn, or deposit operations, you need to connect your wallet."
      />
    );
  }

  return (
    <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
      <TabsList className="grid w-full grid-cols-3 bg-secondary">
        <TabsTrigger
          value="deposit"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          <PiggyBank className="h-4 w-4" />
          Deposit
        </TabsTrigger>
        <TabsTrigger
          value="mint"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          <Coins className="h-4 w-4" />
          Mint
        </TabsTrigger>
        <TabsTrigger
          value="burn"
          className="flex items-center gap-2 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground hover:cursor-pointer"
        >
          <Flame className="h-4 w-4" />
          Burn
        </TabsTrigger>
      </TabsList>

      <TabsContent value="deposit">
        <Card className="border-border bg-card">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2 text-foreground">
                  <PiggyBank className="h-5 w-5 text-primary" />
                  Deposit Collateral
                </CardTitle>
                <CardDescription>
                  Deposit your assets as collateral to mint stablecoins
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRefresh}
                disabled={isLoadingBalance || isLoadingEngine}
              >
                <RefreshCw
                  className={`h-4 w-4 ${
                    isLoadingBalance || isLoadingEngine ? "animate-spin" : ""
                  }`}
                />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="collateral-asset">Collateral Asset</Label>
              <Select value={selectedAsset} onValueChange={setSelectedAsset}>
                <SelectTrigger
                  id="collateral-asset"
                  className="bg-secondary hover:cursor-pointer"
                >
                  <SelectValue placeholder="Select an asset" />
                </SelectTrigger>
                <SelectContent>
                  {collateralAssets.map((asset) => (
                    <SelectItem key={asset.symbol} value={asset.symbol}>
                      <div className="flex items-center gap-2 hover:cursor-pointer">
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
              <div className="flex items-center justify-between text-xs">
                <span className="text-muted-foreground">
                  Available balance:
                </span>
                <span className="font-mono">
                  {isLoadingBalance ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : (
                    `${parseFloat(assetBalance).toFixed(4)} ${selectedAsset}`
                  )}
                </span>
              </div>
            </div>
            {healthFactorProjection && (
              <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
                <div className="flex items-center gap-2 text-sm">
                  <ArrowDownUp className="h-4 w-4 text-primary" />
                  <span className="text-muted-foreground">
                    Health Factor after deposit:
                  </span>
                  <span className="font-mono font-semibold text-primary">
                    {parseFloat(healthFactorProjection).toFixed(2)}
                  </span>
                </div>
              </div>
            )}
            <Button
              className="w-full hover:cursor-pointer"
              size="lg"
              onClick={() => handleSubmit("deposit")}
              disabled={
                !amount ||
                isLoading ||
                parseFloat(amount) > parseFloat(assetBalance)
              }
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
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2 text-foreground">
                  <Coins className="h-5 w-5 text-primary" />
                  Mint Stablecoin
                </CardTitle>
                <CardDescription>
                  Mint stablecoins using your deposited collateral
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRefresh}
                disabled={isLoadingEngine}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isLoadingEngine ? "animate-spin" : ""}`}
                />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border border-border bg-secondary/50 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Available Collateral Value
                </span>
                <span className="font-mono text-foreground">
                  {isLoadingEngine ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : (
                    `$${engineData?.collateral_value_usd
                        ? formatFromWeiPrecise(engineData.collateral_value_usd, 18, 2)
                        : "0.00"
                    }`
                  )}
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Maximum Mintable
                </span>
                <span className="font-mono text-primary">
                  {isLoadingEngine ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : (
                    `$${engineData?.max_mintable
                        ? formatFromWeiPrecise(engineData.max_mintable, 18, 2)
                        : "0.00"
                    }`
                  )}
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Current Health Factor
                </span>
                <span className="font-mono text-foreground">
                  {isLoadingEngine ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : engineData?.current_health_factor ? (
                    parseFloat(formatFromWei(engineData.current_health_factor)).toFixed(2)
                  ) : (
                    "N/A"
                  )}
                </span>
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
                  AUSD
                </span>
              </div>
            </div>
            {healthFactorProjection && (
              <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
                <div className="flex items-center gap-2 text-sm">
                  <ArrowDownUp className="h-4 w-4 text-primary" />
                  <span className="text-muted-foreground">
                    Health Factor after mint:
                  </span>
                  <span className="font-mono font-semibold text-primary">
                    {parseFloat(healthFactorProjection).toFixed(2)}
                  </span>
                </div>
              </div>
            )}
            <Button
              className="w-full"
              size="lg"
              onClick={() => handleSubmit("mint")}
              disabled={
                !amount ||
                isLoading ||
                !engineData ||
                parseFloat(amount) > parseFloat(formatFromWei(engineData.max_mintable || "0"))
              }
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <Coins className="mr-2 h-4 w-4" />
                  Mint AUSD
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="burn">
        <Card className="border-border bg-card">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2 text-foreground">
                  <Flame className="h-5 w-5 text-destructive" />
                  Burn Stablecoin
                </CardTitle>
                <CardDescription>
                  Burn stablecoins to release your collateral
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleRefresh}
                disabled={isLoadingEngine}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isLoadingEngine ? "animate-spin" : ""}`}
                />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border border-border bg-secondary/50 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  AUSD Balance
                </span>
                <span className="font-mono text-foreground">
                  {isLoadingBalance ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : (
                    `${parseFloat(ausdBalance).toFixed(2)} AUSD`
                  )}
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Total Debt
                </span>
                <span className="font-mono text-destructive">
                  {isLoadingEngine ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : (
                    `${
                      engineData?.total_debt
                        ? formatFromWeiPrecise(engineData.total_debt, 18, 2)
                        : "0.00"
                    } AUSD`
                  )}
                </span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  Current Health Factor
                </span>
                <span className="font-mono text-foreground">
                  {isLoadingEngine ? (
                    <Loader2 className="h-3 w-3 animate-spin inline" />
                  ) : engineData?.current_health_factor ? (
                    parseFloat(formatFromWei(engineData.current_health_factor)).toFixed(2)
                  ) : (
                    "N/A"
                  )}
                </span>
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
                  AUSD
                </span>
              </div>
            </div>
            {healthFactorProjection && (
              <div className="rounded-lg border border-primary/20 bg-primary/5 p-3">
                <div className="flex items-center gap-2 text-sm">
                  <ArrowDownUp className="h-4 w-4 text-primary" />
                  <span className="text-muted-foreground">
                    Health Factor after burn:
                  </span>
                  <span className="font-mono font-semibold text-primary">
                    {parseFloat(healthFactorProjection).toFixed(2)}
                  </span>
                </div>
              </div>
            )}
            <Button
              variant="destructive"
              className="w-full"
              size="lg"
              onClick={() => handleSubmit("burn")}
              disabled={
                !amount ||
                isLoading ||
                parseFloat(amount) > parseFloat(ausdBalance) ||
                !engineData ||
                parseFloat(amount) > parseFloat(formatFromWei(engineData.total_debt || "0"))
              }
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <Flame className="mr-2 h-4 w-4" />
                  Burn AUSD
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  );
}
