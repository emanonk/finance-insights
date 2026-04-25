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

export async function getSpendByPrimaryTag(): Promise<TagSpendResponse> {
  const res = await fetch("/api/reports/spend-by-primary-tag");
  if (!res.ok) throw new Error(`Failed to load primary tag report: ${res.status}`);
  return res.json() as Promise<TagSpendResponse>;
}

export async function getSpendBySecondaryTag(): Promise<TagSpendResponse> {
  const res = await fetch("/api/reports/spend-by-secondary-tag");
  if (!res.ok) throw new Error(`Failed to load secondary tag report: ${res.status}`);
  return res.json() as Promise<TagSpendResponse>;
}

export async function getMerchantsByMonth(): Promise<MerchantsByMonthResponse> {
  const res = await fetch("/api/reports/merchants-by-month");
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

export async function getDailySpend(): Promise<DailySpendResponse> {
  const res = await fetch("/api/reports/daily-spend");
  if (!res.ok) throw new Error(`Failed to load daily spend: ${res.status}`);
  return res.json() as Promise<DailySpendResponse>;
}
