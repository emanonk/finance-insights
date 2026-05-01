import { useEffect, useRef, useState } from "react";
import { type Transaction } from "../api/transactions";
import { type Merchant, upsertMerchant } from "../api/merchants";
import { penceToDisplay } from "../lib/money";

interface Props {
  transactions: Transaction[];
  merchants: Map<string, Merchant>;
  onMerchantSaved: (merchant: Merchant) => void;
}

const PAGE_LIMIT = 50;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function amountClass(direction: string): string {
  return direction === "credit"
    ? "text-emerald-600 font-medium"
    : "text-red-500 font-medium";
}

function formatAmount(amount: number, direction: string): string {
  const prefix = direction === "credit" ? "+" : "-";
  return `${prefix}£${penceToDisplay(amount)}`;
}

// ─── Tag chip input ───────────────────────────────────────────────────────────

interface TagChipInputProps {
  tags: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
}

function TagChipInput({ tags, onChange, placeholder }: TagChipInputProps) {
  const [input, setInput] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  function commit(value: string) {
    const trimmed = value.trim();
    if (trimmed && !tags.includes(trimmed)) onChange([...tags, trimmed]);
    setInput("");
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      commit(input);
    } else if (e.key === "Backspace" && input === "" && tags.length > 0) {
      onChange(tags.slice(0, -1));
    }
  }

  return (
    <div
      className="flex flex-wrap gap-1.5 min-h-10 rounded-lg border border-gray-300 bg-white px-2 py-1.5 cursor-text focus-within:ring-2 focus-within:ring-indigo-500 focus-within:border-indigo-500"
      onClick={() => inputRef.current?.focus()}
    >
      {tags.map((tag, i) => (
        <span key={i} className="inline-flex items-center gap-1 rounded-md bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700">
          {tag}
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onChange(tags.filter((_, j) => j !== i)); }}
            className="text-indigo-400 hover:text-indigo-700 leading-none"
          >×</button>
        </span>
      ))}
      <input
        ref={inputRef}
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => { if (input.trim()) commit(input); }}
        placeholder={tags.length === 0 ? placeholder : ""}
        className="flex-1 min-w-24 bg-transparent text-sm text-gray-800 placeholder:text-gray-400 outline-none"
      />
    </div>
  );
}

// ─── Tag modal ────────────────────────────────────────────────────────────────

interface TagModalProps {
  identifier: string;
  existing: Merchant | null;
  onClose: () => void;
  onSaved: (merchant: Merchant) => void;
}

function TagModal({ identifier, existing, onClose, onSaved }: TagModalProps) {
  const [primaryTag, setPrimaryTag] = useState(existing?.primaryTag.name ?? "");
  const [secondaryTags, setSecondaryTags] = useState<string[]>(
    existing?.secondaryTags.map((t) => t.name) ?? []
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    function onKey(e: KeyboardEvent) { if (e.key === "Escape") onClose(); }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function handleSave() {
    if (!primaryTag.trim()) { setError("A primary tag is required."); return; }
    setSaving(true);
    setError(null);
    try {
      const saved = await upsertMerchant(identifier, primaryTag.trim(), secondaryTags);
      onSaved(saved);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unexpected error");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" role="dialog" aria-modal="true">
      <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-2xl bg-white shadow-2xl ring-1 ring-gray-200">
        <div className="flex items-start justify-between border-b border-gray-100 px-6 py-4">
          <div>
            <h2 className="text-base font-semibold text-gray-900">Assign tags</h2>
            <p className="mt-0.5 text-xs text-gray-500 font-mono">{identifier}</p>
          </div>
          <button onClick={onClose} className="ml-4 rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors">
            <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
            </svg>
          </button>
        </div>
        <div className="px-6 py-5 space-y-5">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">
              Primary tag <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={primaryTag}
              onChange={(e) => setPrimaryTag(e.target.value)}
              placeholder="e.g. groceries"
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-800 placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Secondary tags</label>
            <TagChipInput tags={secondaryTags} onChange={setSecondaryTags} placeholder="Type a tag and press Enter" />
            <p className="mt-1 text-xs text-gray-400">Press Enter or comma to add each tag.</p>
          </div>
          {error && (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>
          )}
        </div>
        <div className="flex justify-end gap-3 border-t border-gray-100 px-6 py-4">
          <button onClick={onClose} className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
            Cancel
          </button>
          <button onClick={handleSave} disabled={saving} className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Expanded detail panel ────────────────────────────────────────────────────

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  if (value === null || value === undefined || value === "") return null;
  return (
    <div className="flex gap-2 min-w-0">
      <span className="shrink-0 w-36 text-xs font-semibold text-gray-400 uppercase tracking-wide">{label}</span>
      <span className="text-xs text-gray-700 font-mono break-all">{value}</span>
    </div>
  );
}

function ExpandedPanel({ t }: { t: Transaction }) {
  return (
    <div className="bg-gray-50 border-t border-gray-100 px-6 py-4 space-y-2">
      <DetailRow label="ID" value={t.id} />
      <DetailRow label="Account" value={t.accountId} />
      <DetailRow label="Date" value={t.date} />
      <DetailRow label="Direction" value={t.direction} />
      <DetailRow label="Amount (pence)" value={String(t.amount)} />
      <DetailRow label="Balance before" value={`£${penceToDisplay(t.balanceBefore)}`} />
      <DetailRow label="Balance after" value={`£${penceToDisplay(t.balanceAfter)}`} />
      <DetailRow label="Bank reference" value={t.bankReference} />
      <DetailRow label="Tx reference" value={t.transactionReference} />
      <DetailRow label="Merchant" value={t.merchantIdentifier} />
      <DetailRow label="Statement file" value={t.statementFileName} />
      {t.rawData && t.rawData.length > 0 && (
        <div className="flex gap-2 min-w-0">
          <span className="shrink-0 w-36 text-xs font-semibold text-gray-400 uppercase tracking-wide">Raw data</span>
          <div className="space-y-0.5">
            {t.rawData.map((line, i) => (
              <div key={i} className="text-xs text-gray-700 font-mono break-all">{line}</div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

export function TransactionTable({ transactions, merchants, onMerchantSaved }: Props) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [tagModal, setTagModal] = useState<{ identifier: string; existing: Merchant | null } | null>(null);

  function toggleExpand(id: string) {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  if (transactions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-gray-400">
        <svg xmlns="http://www.w3.org/2000/svg" className="mb-4 h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 17v-2a4 4 0 014-4h4M3 7h18M3 12h6" />
        </svg>
        <p className="text-sm">No transactions found</p>
      </div>
    );
  }

  return (
    <>
      <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="w-8 px-3 py-3" />
              <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">Date</th>
              <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">Bank Reference</th>
              <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">Transaction Reference</th>
              <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">Merchant</th>
              <th className="px-4 py-3 text-right font-semibold text-gray-500 uppercase tracking-wide text-xs">Amount</th>
              <th className="px-4 py-3 text-right font-semibold text-gray-500 uppercase tracking-wide text-xs">Balance</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {transactions.map((t) => {
              const merchant = t.merchantIdentifier ? merchants.get(t.merchantIdentifier) ?? null : null;
              const isExpanded = expandedIds.has(t.id);

              return (
                <>
                  <tr key={t.id} className={`transition-colors ${isExpanded ? "bg-indigo-50/40" : "hover:bg-gray-50"}`}>
                    {/* Expand chevron */}
                    <td className="px-3 py-2 text-center">
                      <button
                        onClick={() => toggleExpand(t.id)}
                        className="text-gray-300 hover:text-gray-500 transition-colors"
                        aria-label={isExpanded ? "Collapse" : "Expand"}
                      >
                        <svg className={`h-4 w-4 transition-transform ${isExpanded ? "rotate-90" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                        </svg>
                      </button>
                    </td>

                    <td className="px-4 py-3 whitespace-nowrap text-gray-500 tabular-nums">{t.date}</td>

                    <td className="px-4 py-3 text-gray-800 max-w-xs truncate">{t.bankReference || "—"}</td>

                    <td className="px-4 py-3 text-gray-800 max-w-xs truncate">{t.transactionReference || "—"}</td>

                    {/* Merchant cell with tags + action button */}
                    <td className="px-4 py-2 max-w-xs">
                      {t.merchantIdentifier ? (
                        <div>
                          <div className="flex items-center gap-1.5">
                            <span className="font-mono text-xs text-gray-800 truncate">{t.merchantIdentifier}</span>
                            <button
                              onClick={() => setTagModal({ identifier: t.merchantIdentifier!, existing: merchant })}
                              className="shrink-0 rounded p-0.5 text-gray-300 hover:text-indigo-600 hover:bg-indigo-50 transition-colors"
                              aria-label={merchant ? "Edit tags" : "Add tags"}
                              title={merchant ? "Edit tags" : "Add tags"}
                            >
                              {merchant ? (
                                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536M9 13l6.586-6.586a2 2 0 112.828 2.828L11.828 15.828a2 2 0 01-1.414.586H8v-2.414a2 2 0 01.586-1.414z" />
                                </svg>
                              ) : (
                                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                                </svg>
                              )}
                            </button>
                          </div>
                          {merchant && (
                            <div className="mt-1 flex flex-wrap gap-1">
                              <span className="inline-flex items-center rounded-md bg-indigo-100 px-1.5 py-0.5 text-xs font-medium text-indigo-700">
                                {merchant.primaryTag.name}
                              </span>
                              {merchant.secondaryTags.map((tag) => (
                                <span key={tag.id} className="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-gray-600">
                                  {tag.name}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>

                    <td className={`px-4 py-3 whitespace-nowrap text-right tabular-nums ${amountClass(t.direction)}`}>
                      {formatAmount(t.amount, t.direction)}
                    </td>

                    <td className="px-4 py-3 whitespace-nowrap text-right text-gray-500 tabular-nums">
                      £{penceToDisplay(t.balanceAfter)}
                    </td>
                  </tr>

                  {isExpanded && (
                    <tr key={`${t.id}-detail`}>
                      <td colSpan={7} className="p-0">
                        <ExpandedPanel t={t} />
                      </td>
                    </tr>
                  )}
                </>
              );
            })}
          </tbody>
        </table>
      </div>

      {tagModal && (
        <TagModal
          identifier={tagModal.identifier}
          existing={tagModal.existing}
          onClose={() => setTagModal(null)}
          onSaved={(m) => { onMerchantSaved(m); setTagModal(null); }}
        />
      )}
    </>
  );
}

export { PAGE_LIMIT };
