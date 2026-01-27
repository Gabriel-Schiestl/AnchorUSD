import { useState, useEffect } from "react";
import {
  useAccount,
  useWriteContract,
  useWaitForTransactionReceipt,
  useReadContract,
  usePublicClient,
} from "wagmi";
import { parseUnits, type Address } from "viem";
import { AUSD_ENGINE_ADDRESS, TOKEN_ADDRESSES } from "@/lib/constants";
import AUSDEngineABI from "@/lib/AUSDEngine.abi.json";
import ERC20ABI from "@/lib/erc20.abi.json";

type OperationStep =
  | "idle"
  | "approving"
  | "approved"
  | "executing"
  | "confirming"
  | "confirmed";

export function useAUSDEngineOperations() {
  const { address } = useAccount();
  const publicClient = usePublicClient();
  const [currentOperation, setCurrentOperation] = useState<string>("");
  const [operationStep, setOperationStep] = useState<OperationStep>("idle");
  const [pendingApprovalHash, setPendingApprovalHash] = useState<
    `0x${string}` | undefined
  >();

  const {
    writeContract,
    data: hash,
    isPending: isWritePending,
    error: writeError,
    reset,
  } = useWriteContract();

  const { isLoading: isConfirming, isSuccess: isConfirmed } =
    useWaitForTransactionReceipt({
      hash,
    });

  const { isSuccess: isApprovalConfirmed } = useWaitForTransactionReceipt({
    hash: pendingApprovalHash,
  });
  useEffect(() => {
    if (isConfirmed) {
      setOperationStep("confirmed");
      setCurrentOperation("");
      setPendingApprovalHash(undefined);
      setTimeout(() => {
        setOperationStep("idle");
        reset();
      }, 2000);
    }
  }, [isConfirmed, reset]);

  const checkAllowance = async (
    tokenAddress: Address,
    amount: string,
    decimals: number = 18,
  ): Promise<boolean> => {
    if (!address || !publicClient) return false;

    try {
      const amountWei = parseUnits(amount, decimals);

      const allowance = (await publicClient.readContract({
        address: tokenAddress,
        abi: ERC20ABI,
        functionName: "allowance",
        args: [address, AUSD_ENGINE_ADDRESS],
      })) as bigint;

      return BigInt(allowance) >= amountWei;
    } catch (error) {
      console.error("Error checking allowance via publicClient:", error);
      return false;
    }
  };

  const approveToken = async (
    tokenAddress: Address,
    amount: string,
    decimals: number = 18,
  ): Promise<void> => {
    setOperationStep("approving");
    setCurrentOperation("Approving token...");

    const amountWei = parseUnits(amount, decimals);

    writeContract(
      {
        address: tokenAddress,
        abi: ERC20ABI,
        functionName: "approve",
        args: [AUSD_ENGINE_ADDRESS, amountWei],
      },
      {
        onSuccess: (txHash) => {
          setPendingApprovalHash(txHash);
          setOperationStep("approved");
          setCurrentOperation("Approval sent, please wait for confirmation...");
        },
        onError: (error) => {
          console.error("Error in approval:", error);
          setOperationStep("idle");
          setCurrentOperation("");
          throw error;
        },
      },
    );
  };

  const depositCollateral = async (
    tokenAddress: Address,
    amount: string,
    decimals: number = 18,
  ) => {
    if (!address) throw new Error("Wallet not connected");

    try {
      setCurrentOperation("Checking approval...");

      const hasAllowance = await checkAllowance(tokenAddress, amount, decimals);
      console.log("depositCollateral -> hasAllowance:", hasAllowance);
      if (!hasAllowance) {
        await approveToken(tokenAddress, amount, decimals);
        return;
      }

      setOperationStep("executing");
      setCurrentOperation("Depositing collateral...");
      const amountWei = parseUnits(amount, decimals);

      try {
        const simulation = await publicClient!.simulateContract({
          address: AUSD_ENGINE_ADDRESS,
          abi: AUSDEngineABI,
          functionName: "depositCollateral",
          args: [tokenAddress, amountWei],
          account: address,
        });

        console.log("Simulation of depositCollateral ok:", simulation);

        writeContract({
          address: AUSD_ENGINE_ADDRESS,
          abi: AUSDEngineABI,
          functionName: "depositCollateral",
          args: [tokenAddress, amountWei],
        });
      } catch (simErr: any) {
        console.error("Simulation reverted:", simErr);
        const reason =
          simErr?.cause?.error?.message ||
          simErr?.shortMessage ||
          simErr?.message;
        throw new Error(
          `Simulation failed before sending: ${reason || String(simErr)}`,
        );
      }
    } catch (error) {
      setOperationStep("idle");
      setCurrentOperation("");
      throw error;
    }
  };

  useEffect(() => {
    if (isApprovalConfirmed && operationStep === "approved") {
      setCurrentOperation("Approval confirmed! Continue the deposit.");
      setOperationStep("idle");
      setPendingApprovalHash(undefined);
    }
  }, [isApprovalConfirmed, operationStep]);

  const mintAUSD = async (amount: string) => {
    if (!address) throw new Error("Wallet not connected");

    try {
      setOperationStep("executing");
      setCurrentOperation("Minting AUSD...");

      const amountWei = parseUnits(amount, 18);

      writeContract({
        address: AUSD_ENGINE_ADDRESS,
        abi: AUSDEngineABI,
        functionName: "mintAUSD",
        args: [amountWei],
      });
    } catch (error) {
      setOperationStep("idle");
      setCurrentOperation("");
      throw error;
    }
  };

  const burnAUSD = async (amount: string) => {
    if (!address) throw new Error("Wallet not connected");

    try {
      setCurrentOperation("Checking approval...");

      const hasAllowance = await checkAllowance(
        TOKEN_ADDRESSES.AUSD,
        amount,
        18,
      );

      if (!hasAllowance) {
        await approveToken(TOKEN_ADDRESSES.AUSD, amount, 18);
        return;
      }

      setOperationStep("executing");
      setCurrentOperation("Burning AUSD...");
      const amountWei = parseUnits(amount, 18);

      writeContract({
        address: AUSD_ENGINE_ADDRESS,
        abi: AUSDEngineABI,
        functionName: "burnAUSD",
        args: [amountWei],
      });
    } catch (error) {
      setOperationStep("idle");
      setCurrentOperation("");
      throw error;
    }
  };

  const redeemCollateral = async (
    tokenAddress: Address,
    amount: string,
    decimals: number = 18,
  ) => {
    if (!address) throw new Error("Wallet not connected");

    try {
      setOperationStep("executing");
      setCurrentOperation("Redeeming collateral...");
      const amountWei = parseUnits(amount, decimals);

      writeContract({
        address: AUSD_ENGINE_ADDRESS,
        abi: AUSDEngineABI,
        functionName: "redeemCollateral",
        args: [tokenAddress, amountWei],
      });
    } catch (error) {
      setOperationStep("idle");
      setCurrentOperation("");
      throw error;
    }
  };

  const isProcessing =
    operationStep !== "idle" && operationStep !== "confirmed";

  return {
    depositCollateral,
    mintAUSD,
    burnAUSD,
    redeemCollateral,
    isProcessing,
    isWritePending,
    isConfirming,
    isConfirmed,
    currentOperation,
    operationStep,
    transactionHash: hash,
    error: writeError,
    needsApproval: operationStep === "approved",
  };
}
