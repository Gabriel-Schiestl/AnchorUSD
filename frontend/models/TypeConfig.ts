import { AlertTriangle, Coins, Flame, PiggyBank } from "lucide-react";

export const typeConfig = {
  deposit: {
    icon: PiggyBank,
    label: "Deposit",
    color: "bg-primary/10 text-primary",
  },
  mint: { icon: Coins, label: "Mint", color: "bg-chart-2/10 text-chart-2" },
  burn: {
    icon: Flame,
    label: "Burn",
    color: "bg-destructive/10 text-destructive",
  },
  liquidation: {
    icon: AlertTriangle,
    label: "Liquidation",
    color: "bg-chart-5/10 text-chart-5",
  },
};
