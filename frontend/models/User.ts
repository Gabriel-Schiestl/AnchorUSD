export interface LiquidatableUser {
  address: string;
  healthFactor: number;
  collateralUsd: string;
  debtUsd: string;
  liquidationAmount: string;
}
