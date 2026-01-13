import { HistoryData } from "@/api/get";

export const mockHistoryData: HistoryData = {
  deposits: [
    {
      id: "1",
      type: "deposit",
      amount: "2.5",
      asset: "ETH",
      timestamp: "2025-01-12T10:30:00Z",
      txHash: "0x1234...abcd",
      status: "completed",
    },
    {
      id: "2",
      type: "deposit",
      amount: "0.08",
      asset: "WBTC",
      timestamp: "2025-01-11T14:20:00Z",
      txHash: "0x5678...efgh",
      status: "completed",
    },
    {
      id: "3",
      type: "deposit",
      amount: "1.2",
      asset: "ETH",
      timestamp: "2025-01-10T09:15:00Z",
      txHash: "0x9abc...ijkl",
      status: "completed",
    },
  ],
  mintBurn: [
    {
      id: "4",
      type: "mint",
      amount: "5,000.00",
      timestamp: "2025-01-12T11:00:00Z",
      txHash: "0xdef0...mnop",
      status: "completed",
    },
    {
      id: "5",
      type: "burn",
      amount: "1,500.00",
      timestamp: "2025-01-11T16:45:00Z",
      txHash: "0x1111...qrst",
      status: "completed",
    },
    {
      id: "6",
      type: "mint",
      amount: "3,500.00",
      timestamp: "2025-01-09T08:30:00Z",
      txHash: "0x2222...uvwx",
      status: "completed",
    },
  ],
  liquidations: [
    {
      id: "7",
      type: "liquidation",
      amount: "800.00",
      timestamp: "2025-01-05T22:10:00Z",
      txHash: "0x3333...yzab",
      status: "completed",
    },
  ],
} as const;
