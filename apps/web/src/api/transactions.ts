export interface Transaction {
  id: string;
  accountId: string;
  date: string;
  bankReference: string | null;
  transactionReference: string | null;
  merchantIdentifier: string | null;
  balanceBefore: number;
  balanceAfter: number;
  amount: number;
  direction: string;
  rawData: string[] | null;
  statementFileName: string | null;
}

export interface TransactionListResponse {
  items: Transaction[];
  total: number;
  limit: number;
  offset: number;
}

export async function listTransactions(
  limit: number,
  offset: number,
  accountIds?: string[]
): Promise<TransactionListResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  if (accountIds && accountIds.length > 0) {
    params.set("accountIds", accountIds.join(","));
  }
  const res = await fetch(`/api/transactions?${params}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch transactions: ${res.status}`);
  }
  return res.json() as Promise<TransactionListResponse>;
}
