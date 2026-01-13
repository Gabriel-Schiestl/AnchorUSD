import { get } from "./get";
import { post } from "./post";

export interface AUSDEngineData {
  ausdBalance: string;
  totalDebt: string;
  collateralValueUSD: string;
  maxMintable: string;
  currentHealthFactor: string;
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

export const ausdEngineApi = {
  getUserData: async (address: string): Promise<AUSDEngineData> => {
    return get<AUSDEngineData>(`/api/ausd-engine/user/${address}`);
  },

  calculateHealthFactorAfterMint: async (
    address: string,
    mintAmount: string
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateMintRequest>(
      `/api/ausd-engine/calculate-mint`,
      { address, mintAmount }
    );
  },

  calculateHealthFactorAfterBurn: async (
    address: string,
    burnAmount: string
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateBurnRequest>(
      `/api/ausd-engine/calculate-burn`,
      { address, burnAmount }
    );
  },

  calculateHealthFactorAfterDeposit: async (
    address: string,
    tokenAddress: string,
    depositAmount: string
  ): Promise<HealthFactorProjection> => {
    return post<HealthFactorProjection, CalculateDepositRequest>(
      `/api/ausd-engine/calculate-deposit`,
      { address, tokenAddress, depositAmount }
    );
  },
};
