import { useCallback, useEffect, useRef, useState } from "react";
import {
  getTopIdentifiers,
  upsertMerchant,
  type IdentifierCount,
  type Merchant,
} from "../api/merchants";

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
    if (trimmed && !tags.includes(trimmed)) {
      onChange([...tags, trimmed]);
    }
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

  function removeTag(index: number) {
    onChange(tags.filter((_, i) => i !== index));
  }

  return (
    <div
      className="flex flex-wrap gap-1.5 min-h-10 rounded-lg border border-gray-300 bg-white px-2 py-1.5 cursor-text focus-within:ring-2 focus-within:ring-indigo-500 focus-within:border-indigo-500"
      onClick={() => inputRef.current?.focus()}
    >
      {tags.map((tag, i) => (
        <span
          key={i}
          className="inline-flex items-center gap-1 rounded-md bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700"
        >
          {tag}
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); removeTag(i); }}
            className="text-indigo-400 hover:text-indigo-700 leading-none"
            aria-label={`Remove ${tag}`}
          >
            ×
          </button>
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

// ─── Tag assignment modal ─────────────────────────────────────────────────────

interface TagModalProps {
  item: IdentifierCount;
  onClose: () => void;
  onSaved: (merchant: Merchant) => void;
}

function TagModal({ item, onClose, onSaved }: TagModalProps) {
  const [primaryTag, setPrimaryTag] = useState(
    item.merchant?.primaryTag.name ?? ""
  );
  const [secondaryTags, setSecondaryTags] = useState<string[]>(
    item.merchant?.secondaryTags.map((t) => t.name) ?? []
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSave() {
    if (!primaryTag.trim()) {
      setError("A primary tag is required.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const saved = await upsertMerchant(
        item.identifier,
        primaryTag.trim(),
        secondaryTags
      );
      onSaved(saved);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unexpected error");
    } finally {
      setSaving(false);
    }
  }

  // Close on Escape
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Assign tags"
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Panel */}
      <div className="relative z-10 w-full max-w-md rounded-2xl bg-white shadow-2xl ring-1 ring-gray-200">
        {/* Header */}
        <div className="flex items-start justify-between border-b border-gray-100 px-6 py-4">
          <div>
            <h2 className="text-base font-semibold text-gray-900">
              Assign tags
            </h2>
            <p className="mt-0.5 text-xs text-gray-500 font-mono">
              {item.identifier}
            </p>
          </div>
          <button
            onClick={onClose}
            className="ml-4 rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
            aria-label="Close"
          >
            <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
            </svg>
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-5 space-y-5">
          {/* Primary tag */}
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

          {/* Secondary tags */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">
              Secondary tags
            </label>
            <TagChipInput
              tags={secondaryTags}
              onChange={setSecondaryTags}
              placeholder="Type a tag and press Enter"
            />
            <p className="mt-1 text-xs text-gray-400">
              Press Enter or comma to add each tag.
            </p>
          </div>

          {error && (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 border-t border-gray-100 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Merchants view ───────────────────────────────────────────────────────────

export function MerchantsView() {
  const [items, setItems] = useState<IdentifierCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<IdentifierCount | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getTopIdentifiers(50);
      setItems(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unexpected error");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  function handleSaved(merchant: Merchant) {
    setItems((prev) =>
      prev.map((ic) =>
        ic.identifier === merchant.identifierName
          ? { ...ic, merchant }
          : ic
      )
    );
    setSelected(null);
  }

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-gray-900">Merchants</h2>
        {!loading && !error && (
          <p className="mt-0.5 text-sm text-gray-500">
            {items.length} most frequent merchant identifiers — click one to assign tags.
          </p>
        )}
      </div>

      {error && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="space-y-2">
          {Array.from({ length: 8 }).map((_, i) => (
            <div
              key={i}
              className="h-14 animate-pulse rounded-xl bg-gray-100"
            />
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-gray-400">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="mb-4 h-12 w-12"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={1}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M3 6h18M3 12h18M3 18h18"
            />
          </svg>
          <p className="text-sm">No merchant identifiers found yet.</p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
          {items.map((ic, idx) => (
            <button
              key={ic.identifier}
              onClick={() => setSelected(ic)}
              className={`w-full flex items-center gap-4 px-5 py-3.5 text-left hover:bg-gray-50 transition-colors ${
                idx < items.length - 1 ? "border-b border-gray-100" : ""
              }`}
            >
              {/* Rank badge */}
              <span className="w-7 shrink-0 text-xs font-semibold text-gray-400 tabular-nums text-right">
                {idx + 1}
              </span>

              {/* Identifier + tags */}
              <div className="flex-1 min-w-0">
                <span className="block truncate text-sm font-medium text-gray-800 font-mono">
                  {ic.identifier}
                </span>
                {ic.merchant ? (
                  <span className="mt-0.5 flex flex-wrap gap-1">
                    <span className="inline-flex items-center rounded-md bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700">
                      {ic.merchant.primaryTag.name}
                    </span>
                    {ic.merchant.secondaryTags.map((t) => (
                      <span
                        key={t.id}
                        className="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600"
                      >
                        {t.name}
                      </span>
                    ))}
                  </span>
                ) : (
                  <span className="mt-0.5 block text-xs text-gray-400 italic">
                    No tags assigned
                  </span>
                )}
              </div>

              {/* Count badge */}
              <span className="shrink-0 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-semibold text-gray-600 tabular-nums">
                {ic.count}×
              </span>

              {/* Edit chevron */}
              <svg
                className="h-4 w-4 shrink-0 text-gray-300"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          ))}
        </div>
      )}

      {selected && (
        <TagModal
          item={selected}
          onClose={() => setSelected(null)}
          onSaved={handleSaved}
        />
      )}
    </>
  );
}
