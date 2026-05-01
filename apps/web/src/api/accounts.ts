export interface Account {
  id: string;
  bankName: string;
  accountNumber: string;
}

export interface AccountListResponse {
  items: Account[];
}

export async function listAccounts(): Promise<AccountListResponse> {
  const res = await fetch("/api/accounts");
  if (!res.ok) throw new Error(`Failed to fetch accounts: ${res.status}`);
  return res.json() as Promise<AccountListResponse>;
}
