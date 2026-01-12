import { MintBurnDeposit } from "@/components/operations/mint-burn-deposit";

export default function HomePage() {
  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">
          Mint / Burn / Deposit
        </h1>
        <p className="mt-2 text-muted-foreground">
          Manage your collateral and stablecoins in one interface
        </p>
      </div>
      <MintBurnDeposit />
    </div>
  );
}
