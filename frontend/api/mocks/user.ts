import { MetricsData } from "../get";

export const mockRiskData: MetricsData = {
  liquidatableUsers: [
    {
      address: "0x1234...5678",
      healthFactor: 0.95,
      collateralUsd: "4,200.00",
      debtUsd: "4,500.00",
      liquidationAmount: "300.00",
    },
    {
      address: "0xabcd...ef01",
      healthFactor: 0.88,
      collateralUsd: "8,500.00",
      debtUsd: "9,800.00",
      liquidationAmount: "1,300.00",
    },
    {
      address: "0x9876...fedc",
      healthFactor: 0.92,
      collateralUsd: "2,100.00",
      debtUsd: "2,300.00",
      liquidationAmount: "200.00",
    },
  ],
  totalCollateral: {
    value: "125,000,000",
    breakdown: [
      {
        asset: "ETH",
        amount: "45,000",
        valueUsd: "85,000,000",
        percentage: 68.4,
      },
      {
        asset: "WBTC",
        amount: "500",
        valueUsd: "32,000,000",
        percentage: 25.6,
      },
    ],
  },
  stableSupply: {
    total: "75,500,000",
    circulating: "72,300,000",
    backing: 165.5,
  },
  protocolHealth: {
    averageHealthFactor: 2.15,
    usersAtRisk: 42,
    totalUsers: 12580,
    collateralizationRatio: 165.5,
  },
};
