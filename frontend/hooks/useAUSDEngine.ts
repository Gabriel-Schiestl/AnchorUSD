import useSWR from "swr";
import { useAccount } from "wagmi";
import {
  ausdEngineApi,
  AUSDEngineData,
  HealthFactorProjection,
} from "@/api/ausd-engine";

function toScaledIntegerString(amount: string, decimals = 18): string {
  if (!amount) return "0";
  const [wholePart, fracPart = ""] = amount.split(".");
  const whole =
    wholePart === "" ? "0" : wholePart.replace(/^0+(?=\d)|\D/g, "") || "0";
  const frac = fracPart.replace(/\D/g, "");
  const paddedFrac = (frac + "0".repeat(decimals)).slice(0, decimals);
  const multiplier = BigInt("1" + "0".repeat(decimals));
  const wholeBig = BigInt(whole);
  const fracBig = BigInt(paddedFrac || "0");
  return (wholeBig * multiplier + fracBig).toString();
}

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
    },
  );

  const calculateHealthFactorAfterMint = async (
    mintAmount: string,
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    const scaledMintAmount = toScaledIntegerString(mintAmount, 18);

    try {
      return await ausdEngineApi.calculateHealthFactorAfterMint(
        address,
        scaledMintAmount.toString(),
      );
    } catch (err) {
      console.error("Error calculating health factor after mint:", err);
      return null;
    }
  };

  const calculateHealthFactorAfterBurn = async (
    burnAmount: string,
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    try {
      return await ausdEngineApi.calculateHealthFactorAfterBurn(
        address,
        toScaledIntegerString(burnAmount, 18),
      );
    } catch (err) {
      console.error("Error calculating health factor after burn:", err);
      return null;
    }
  };

  const calculateHealthFactorAfterDeposit = async (
    tokenAddress: string,
    depositAmount: string,
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    const scaledDepositAmount = toScaledIntegerString(depositAmount, 18);

    try {
      return await ausdEngineApi.calculateHealthFactorAfterDeposit(
        address,
        tokenAddress,
        scaledDepositAmount.toString(),
      );
    } catch (err) {
      console.error("Error calculating health factor after deposit:", err);
      return null;
    }
  };

  const calculateHealthFactorAfterRedeem = async (
    tokenAddress: string,
    redeemAmount: string,
  ): Promise<HealthFactorProjection | null> => {
    if (!address) return null;

    const scaledRedeemAmount = toScaledIntegerString(redeemAmount, 18);

    try {
      return await ausdEngineApi.calculateHealthFactorAfterRedeem(
        address,
        tokenAddress,
        scaledRedeemAmount.toString(),
      );
    } catch (err) {
      console.error("Error calculating health factor after redeem:", err);
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
    calculateHealthFactorAfterRedeem,
  };
}
