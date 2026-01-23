import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { formatUnits } from "viem";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Convert wei (18 decimals) to human readable format using viem
 * @param wei - Amount in wei as string
 * @param decimals - Number of decimals (default 18)
 * @returns Formatted string
 */
export function formatFromWei(wei: string, decimals: number = 18): string {
  try {
    return formatUnits(BigInt(wei), decimals);
  } catch (error) {
    console.error("Error formatting wei:", error);
    return "0";
  }
}

/**
 * Format wei with custom precision (removes trailing zeros)
 * @param wei - Amount in wei as string
 * @param decimals - Number of decimals (default 18)
 * @param maxDecimals - Maximum decimal places to show (default 4)
 * @returns Formatted string
 */
export function formatFromWeiPrecise(
  wei: string,
  decimals: number = 18,
  maxDecimals: number = 4,
): string {
  try {
    const formatted = formatUnits(BigInt(wei), decimals);
    const [integer, decimal] = formatted.split(".");

    if (!decimal) {
      return integer;
    }

    const truncated = decimal.slice(0, maxDecimals).replace(/0+$/, "");

    if (truncated.length === 0) {
      return integer;
    }

    return `${integer}.${truncated}`;
  } catch (error) {
    console.error("Error formatting wei:", error);
    return "0";
  }
}
