export function getHealthFactorColor(hf: number) {
  if (hf >= 2) return "text-primary";
  if (hf >= 1.5) return "text-chart-3";
  if (hf >= 1.2) return "text-chart-5";
  return "text-destructive";
}

export function getHealthFactorStatus(hf: number) {
  if (hf >= 2) return "Healthy";
  if (hf >= 1.5) return "Moderate";
  if (hf >= 1.2) return "At Risk";
  return "Critical";
}

export function getHealthPercent(hf: number) {
  return Math.min((hf / 3) * 100, 100);
}
