export interface Tag {
  id: number;
  name: string;
  type: "primary" | "secondary";
}

export interface Merchant {
  id: number;
  identifierName: string;
  primaryTag: Tag;
  secondaryTags: Tag[];
}

export interface IdentifierCount {
  identifier: string;
  count: number;
  merchant: Merchant | null;
}

export interface TopIdentifiersResponse {
  items: IdentifierCount[];
}

export async function getTopIdentifiers(limit = 30): Promise<TopIdentifiersResponse> {
  const res = await fetch(`/api/merchants/top?limit=${limit}`);
  if (!res.ok) throw new Error(`Failed to fetch top identifiers: ${res.status}`);
  return res.json() as Promise<TopIdentifiersResponse>;
}

export async function upsertMerchant(
  identifierName: string,
  primaryTagName: string,
  secondaryTagNames: string[]
): Promise<Merchant> {
  const res = await fetch("/api/merchants", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ identifierName, primaryTagName, secondaryTagNames }),
  });
  if (!res.ok) throw new Error(`Failed to save merchant: ${res.status}`);
  return res.json() as Promise<Merchant>;
}
