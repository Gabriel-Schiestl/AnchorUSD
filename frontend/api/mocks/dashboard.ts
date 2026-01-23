export interface CollateralDeposited {
  asset: string;
  amount: string;
  valueUsd: string;
}

export interface DashboardData {
  total_debt: string;
  collateral_value_usd: string;
  max_mintable: string;
  current_health_factor: string;
  collateral_deposited: CollateralDeposited[];
}

export const mockDashboardData: DashboardData = {
  total_debt: "6500000000000000000000", // 6500 AUSD in wei
  collateral_value_usd: "18800000000000000000000", // 18800 USD in wei (scaled)
  max_mintable: "5000000000000000000000", // 5000 AUSD in wei
  current_health_factor: "1850000000000000000", // 1.85 (scaled with 18 decimals)
  collateral_deposited: [
    {
      asset: "ETH",
      amount: "5250000000000000000", // 5.25 ETH in wei
      valueUsd: "12500000000000000000000", // 12500 USD (scaled)
    },
    {
      asset: "BTC",
      amount: "150000000000000000", // 0.15 BTC in wei
      valueUsd: "6300000000000000000000", // 6300 USD (scaled)
    },
  ],
};
