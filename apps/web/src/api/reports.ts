export interface TagSpend {
  tagName: string;
  total: string;
  count: number;
}

export interface TagSpendResponse {
  items: TagSpend[];
}

export interface MerchantMonthly {
  month: string;
  total: string;
  maxAmount: string;
  avgAmount: string;
  count: number;
}

export interface MerchantSummary {
  identifier: string;
  totalSpend: string;
  txCount: number;
  months: MerchantMonthly[];
}

export interface MerchantsByMonthResponse {
  items: MerchantSummary[];
}

function accountParam(accountIds?: string[]): string {
  if (!accountIds || accountIds.length === 0) return "";
  return `?accountIds=${accountIds.join(",")}`;
}

export async function getSpendByPrimaryTag(accountIds?: string[]): Promise<TagSpendResponse> {
  const res = await fetch(`/api/reports/spend-by-primary-tag${accountParam(accountIds)}`);
  if (!res.ok) throw new Error(`Failed to load primary tag report: ${res.status}`);
  return res.json() as Promise<TagSpendResponse>;
}

export async function getSpendBySecondaryTag(accountIds?: string[]): Promise<TagSpendResponse> {
  const res = await fetch(`/api/reports/spend-by-secondary-tag${accountParam(accountIds)}`);
  if (!res.ok) throw new Error(`Failed to load secondary tag report: ${res.status}`);
  return res.json() as Promise<TagSpendResponse>;
}

export async function getMerchantsByMonth(accountIds?: string[]): Promise<MerchantsByMonthResponse> {
  const res = await fetch(`/api/reports/merchants-by-month${accountParam(accountIds)}`);
  if (!res.ok) throw new Error(`Failed to load merchants report: ${res.status}`);
  return res.json() as Promise<MerchantsByMonthResponse>;
}

export interface DailySpend {
  date: string; // "YYYY-MM-DD"
  total: string;
}

export interface DailySpendResponse {
  items: DailySpend[];
}

export async function getDailySpend(accountIds?: string[]): Promise<DailySpendResponse> {
  const res = await fetch(`/api/reports/daily-spend${accountParam(accountIds)}`);
  if (!res.ok) throw new Error(`Failed to load daily spend: ${res.status}`);
  return res.json() as Promise<DailySpendResponse>;
}
