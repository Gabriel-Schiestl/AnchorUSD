import { get } from "./get";
import { post } from "./post";

export interface CollateralDeposited {
  asset: string;
  amount: string;
  valueUsd: string;
}

export interface AUSDEngineData {
  total_debt: string;
  collateral_value_usd: string;
  max_mintable: string;
  current_health_factor: string;
  collateral_deposited: CollateralDeposited[];
}

export interface HealthFactorProjection {
  healthFactorAfter: string;
  newDebt: string;
  newCollateralValue: string;
}

interface CalculateMintRequest {
  address: string;
  mintAmount: string;
}

interface CalculateBurnRequest {
  address: string;
  burnAmount: string;
}

interface CalculateDepositRequest {
  address: string;
  tokenAddress: string;
  depositAmount: string;
}

interface CalculateRedeemRequest {
  address: string;
  tokenAddress: string;
  redeemAmount: string;
}

export const ausdEngineApi = {
  getUserData: async (address: string): Promise<AUSDEngineData> => {
    return get<AUSDEngineData>(`/api/user/${address}`);
  },

  calculateHealthFactorAfterMint: async (
    address: string,
    mintAmount: string,
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateMintRequest>(
      `/api/ausd-engine/calculate-mint`,
      { address, mintAmount },
    );
  },

  calculateHealthFactorAfterBurn: async (
    address: string,
    burnAmount: string,
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateBurnRequest>(
      `/api/ausd-engine/calculate-burn`,
      { address, burnAmount },
    );
  },

  calculateHealthFactorAfterDeposit: async (
    address: string,
    tokenAddress: string,
    depositAmount: string,
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateDepositRequest>(
      `/api/ausd-engine/calculate-deposit`,
      { address, tokenAddress, depositAmount },
    );
  },

  calculateHealthFactorAfterRedeem: async (
    address: string,
    tokenAddress: string,
    redeemAmount: string,
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateRedeemRequest>(
      `/api/ausd-engine/calculate-redeem`,
      { address, tokenAddress, redeemAmount },
    );
  },
};
