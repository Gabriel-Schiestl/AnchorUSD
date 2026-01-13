// Endereços dos contratos - Ajuste conforme seu deployment
export const TOKEN_ADDRESSES = {
  WETH: (process.env.NEXT_PUBLIC_WETH_ADDRESS ||
    "0x0000000000000000000000000000000000000000") as `0x${string}`,
  WBTC: (process.env.NEXT_PUBLIC_WBTC_ADDRESS ||
    "0x0000000000000000000000000000000000000000") as `0x${string}`,
  AUSD: (process.env.NEXT_PUBLIC_AUSD_ADDRESS ||
    "0x0000000000000000000000000000000000000000") as `0x${string}`,
} as const;

export const collateralAssets = [
  {
    symbol: "WETH",
    name: "Wrapped Ethereum",
    icon: "Ξ",
    address: TOKEN_ADDRESSES.WETH,
  },
  {
    symbol: "WBTC",
    name: "Wrapped Bitcoin",
    icon: "₿",
    address: TOKEN_ADDRESSES.WBTC,
  },
] as const;
