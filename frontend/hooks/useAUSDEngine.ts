import useSWR from "swr";
import { useAccount } from "wagmi";
import {
  ausdEngineApi,
  AUSDEngineData,
  HealthFactorProjection,
} from "@/api/ausd-engine";

export function useAUSDEngine() {
  const { address, isConnected } = useAccount();

  const {
    data: engineData,
    isLoading,
    error,
    mutate,
  } = useSWR<AUSDEngineData>(
    address && isConnected ? `/api/ausd-engine/user/${address}` : null,
    () => ausdEngineApi.getUserData(address!),
    {
      refreshInterval: 10000,
      revalidateOnFocus: true,
      revalidateOnReconnect: true,
    }
  );

  const calculateHealthFactorAfterMint = async (
    mintAmount: string
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    try {
      return await ausdEngineApi.calculateHealthFactorAfterMint(
        address,
        mintAmount
      );
    } catch (err) {
      console.error("Error calculating health factor after mint:", err);
      return null;
    }
  };

  const calculateHealthFactorAfterBurn = async (
    burnAmount: string
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    try {
      return await ausdEngineApi.calculateHealthFactorAfterBurn(
        address,
        burnAmount
      );
    } catch (err) {
      console.error("Error calculating health factor after burn:", err);
      return null;
    }
  };

  const calculateHealthFactorAfterDeposit = async (
    tokenAddress: string,
    depositAmount: string
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    try {
      return await ausdEngineApi.calculateHealthFactorAfterDeposit(
        address,
        tokenAddress,
        depositAmount
      );
    } catch (err) {
      console.error("Error calculating health factor after deposit:", err);
      return null;
    }
  };

  return {
    engineData: engineData || null,
    isLoading,
    error: error?.message || null,
    refresh: mutate,
    calculateHealthFactorAfterMint,
    calculateHealthFactorAfterBurn,
    calculateHealthFactorAfterDeposit,
  };
}
