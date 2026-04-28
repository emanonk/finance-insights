export interface UploadResult {
  fileName: string;
  transactionCount: number;
}

export async function uploadStatement(bankName: string, file: File): Promise<UploadResult> {
  const form = new FormData();
  form.append("bank_name", bankName);
  form.append("file", file);

  const res = await fetch("/api/statements", { method: "POST", body: form });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `Upload failed (${res.status})`);
  }
  return res.json() as Promise<UploadResult>;
}
