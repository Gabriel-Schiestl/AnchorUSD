import { HistoryList } from "@/components/history/history-list";

export default function HistoryPage() {
  return (
    <div className="mx-auto max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">History</h1>
        <p className="mt-2 text-muted-foreground">
          Track all your transactions and events
        </p>
      </div>
      <HistoryList />
    </div>
  );
}
