export type TransactionType = "deposit" | "mint" | "burn" | "liquidation";

export default interface Transaction {
  id: string;
  type: TransactionType;
  amount: string;
  asset?: string;
  timestamp: string;
  txHash: string;
  status: "completed" | "pending" | "failed";
}
