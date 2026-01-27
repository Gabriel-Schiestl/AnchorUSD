import { useEffect, useState } from "react";
import { useAccount, useBalance } from "wagmi";
import { formatUnits } from "viem";

export function useWalletBalance(tokenAddress?: `0x${string}`) {
  console.log("useWalletBalance -> tokenAddress:", tokenAddress);
  const { address, isConnected } = useAccount();
  const [balance, setBalance] = useState<string>("0");
  const [isLoading, setIsLoading] = useState(false);

  const { data: balanceData, refetch } = useBalance({
    address: address,
    token: tokenAddress,
  });

  useEffect(() => {
    if (balanceData) {
      const formattedBalance = formatUnits(
        balanceData.value,
        balanceData.decimals,
      );
      setBalance(formattedBalance);
    } else {
      setBalance("0");
    }
  }, [balanceData]);

  const refresh = async () => {
    setIsLoading(true);
    await refetch();
    setIsLoading(false);
  };

  return {
    balance,
    symbol: balanceData?.symbol || "",
    decimals: balanceData?.decimals || 18,
    isLoading,
    refresh,
    isConnected,
  };
}
