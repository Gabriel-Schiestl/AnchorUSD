export interface DashboardData {
  collateralDeposited: { asset: string; amount: string; valueUsd: string }[];
  collateral: { total: string; locked: string; available: string };
  healthFactor: number;
  debt: string;
}

export const mockDashboardData: DashboardData = {
  collateralDeposited: [
    { asset: "WETH", amount: "5.25", valueUsd: "12,500.00" },
    { asset: "WBTC", amount: "0.15", valueUsd: "6,300.00" },
  ],
  collateral: {
    total: "18,800.00",
    locked: "12,000.00",
    available: "6,800.00",
  },
  healthFactor: 1.85,
  debt: "6,500.00",
};
