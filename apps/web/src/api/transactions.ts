export interface Transaction {
  id: string;
  accountId: number;
  date: string;
  bankReferenceNumber: string | null;
  justification: string | null;
  indicator: string | null;
  merchantIdentifier: string | null;
  amount1: string | null;
  mccCode: string | null;
  cardMasked: string | null;
  reference: string | null;
  description: string;
  paymentMethod: string | null;
  direction: string;
  amount: string;
  balanceAfterTransaction: string | null;
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
  offset: number
): Promise<TransactionListResponse> {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  const res = await fetch(`/api/transactions?${params}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch transactions: ${res.status}`);
  }
  return res.json() as Promise<TransactionListResponse>;
}
