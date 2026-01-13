import { ConnectButton } from "@rainbow-me/rainbowkit";
import { Card, CardContent } from "@/components/ui/card";
import { LucideIcon } from "lucide-react";

interface ConnectWalletPromptProps {
  icon?: LucideIcon;
  title?: string;
  description?: string;
}

export function ConnectWalletPrompt({
  icon: Icon,
  title = "Connect your wallet",
  description = "To perform operations, you need to connect your wallet.",
}: ConnectWalletPromptProps) {
  return (
    <Card className="border-border bg-card">
      <CardContent className="flex flex-col items-center justify-center py-16">
        {Icon && (
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <Icon className="h-8 w-8 text-primary" />
          </div>
        )}
        <h3 className="mb-2 text-xl font-semibold text-foreground">{title}</h3>
        <p className="mb-6 text-center text-muted-foreground">{description}</p>
        <ConnectButton />
      </CardContent>
    </Card>
  );
}
