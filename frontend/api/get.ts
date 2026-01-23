import { axiosInstance } from "@/api/instance";
import Transaction from "@/models/Transaction";

export interface HistoryData {
  deposits: Transaction[];
  mintBurn: Transaction[];
  liquidations: Transaction[];
}

export interface LiquidatableUser {
  address: string;
  healthFactor: string;
  collateralUsd: string;
  debtUsd: string;
  liquidationAmount: string;
}

export interface MetricsData {
  liquidatableUsers: LiquidatableUser[];
  totalCollateral: {
    value: string;
    breakdown: {
      asset: string;
      amount: string;
      valueUsd: string;
      percentage: number;
    }[];
  };
  stableSupply: {
    total: string;
    circulating: string;
    backing: number;
  };
  protocolHealth: {
    averageHealthFactor: number;
    usersAtRisk: number;
    totalUsers: number;
    collateralizationRatio: number;
  };
}

type PathsAccepted = "/history" | "/dashboard" | "/risk" | string;

export const get = async <T>(path: PathsAccepted): Promise<T> => {
  const response = await axiosInstance.get<T>(path);
  return response.data;
};
