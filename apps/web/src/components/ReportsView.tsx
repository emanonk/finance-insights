import { useCallback, useEffect, useState } from "react";
import {
  getSpendByPrimaryTag,
  getSpendBySecondaryTag,
  getMerchantsByMonth,
  type TagSpend,
  type MerchantSummary,
} from "../api/reports";

// ─── Helpers ─────────────────────────────────────────────────────────────────

function fmt(value: string): string {
  const n = parseFloat(value);
  if (isNaN(n)) return value;
  return n.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function Skeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="h-10 animate-pulse rounded-lg bg-gray-100" />
      ))}
    </div>
  );
}

function EmptyState({ label }: { label: string }) {
  return (
    <p className="py-10 text-center text-sm text-gray-400">{label}</p>
  );
}

function SectionError({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {message}
    </div>
  );
}

// ─── Tag spend bar chart ──────────────────────────────────────────────────────

function TagSpendList({ items }: { items: TagSpend[] }) {
  if (items.length === 0) return <EmptyState label="No tagged spending data yet." />;

  const max = Math.max(...items.map((i) => parseFloat(i.total) || 0));

  return (
    <div className="space-y-2.5">
      {items.map((item) => {
        const pct = max > 0 ? ((parseFloat(item.total) || 0) / max) * 100 : 0;
        return (
          <div key={item.tagName} className="flex items-center gap-3">
            <span className="w-32 shrink-0 truncate text-sm font-medium text-gray-700">
              {item.tagName}
            </span>
            <div className="relative flex-1 h-7 rounded-md bg-gray-100 overflow-hidden">
              <div
                className="absolute inset-y-0 left-0 rounded-md bg-indigo-500 transition-all"
                style={{ width: `${pct}%` }}
              />
              <span className="absolute inset-y-0 right-2 flex items-center text-xs font-semibold text-gray-700">
                €{fmt(item.total)}
              </span>
            </div>
            <span className="w-14 shrink-0 text-right text-xs text-gray-400 tabular-nums">
              {item.count}×
            </span>
          </div>
        );
      })}
    </div>
  );
}

// ─── Merchants by month ───────────────────────────────────────────────────────

function MerchantsByMonthTable({ items }: { items: MerchantSummary[] }) {
  const [expanded, setExpanded] = useState<string | null>(null);

  if (items.length === 0) return <EmptyState label="No merchant data yet." />;

  // Collect all unique months across all merchants (sorted).
  const allMonths = Array.from(
    new Set(items.flatMap((m) => m.months.map((mo) => mo.month)))
  ).sort();

  // Show at most the 8 most recent months in the summary row.
  const displayMonths = allMonths.slice(-8);

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="min-w-full text-sm">
        <thead className="bg-gray-50 border-b border-gray-200">
          <tr>
            <th className="sticky left-0 z-10 bg-gray-50 px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 min-w-40">
              Merchant
            </th>
            <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">
              Total
            </th>
            <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">
              Txs
            </th>
            {displayMonths.map((m) => (
              <th
                key={m}
                className="px-3 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 whitespace-nowrap"
              >
                {m}
              </th>
            ))}
            <th className="w-8" />
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {items.map((merchant) => {
            const monthMap = Object.fromEntries(
              merchant.months.map((mo) => [mo.month, mo])
            );
            const isOpen = expanded === merchant.identifier;

            return (
              <>
                <tr
                  key={merchant.identifier}
                  className="hover:bg-gray-50 cursor-pointer transition-colors"
                  onClick={() =>
                    setExpanded(isOpen ? null : merchant.identifier)
                  }
                >
                  <td className="sticky left-0 bg-white px-4 py-3 font-medium text-gray-800 font-mono text-xs whitespace-nowrap">
                    {merchant.identifier}
                  </td>
                  <td className="px-4 py-3 text-right font-semibold text-gray-800 tabular-nums whitespace-nowrap">
                    €{fmt(merchant.totalSpend)}
                  </td>
                  <td className="px-4 py-3 text-right text-gray-500 tabular-nums">
                    {merchant.txCount}
                  </td>
                  {displayMonths.map((m) => {
                    const mo = monthMap[m];
                    return (
                      <td
                        key={m}
                        className="px-3 py-3 text-right text-gray-600 tabular-nums whitespace-nowrap"
                      >
                        {mo ? `€${fmt(mo.total)}` : "—"}
                      </td>
                    );
                  })}
                  <td className="px-3 py-3 text-gray-300">
                    <svg
                      className={`h-4 w-4 transition-transform ${isOpen ? "rotate-90" : ""}`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </td>
                </tr>

                {isOpen && (
                  <tr key={`${merchant.identifier}-detail`}>
                    <td
                      colSpan={4 + displayMonths.length}
                      className="bg-indigo-50 px-4 py-4"
                    >
                      <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-indigo-600">
                        Monthly breakdown — {merchant.identifier}
                      </p>
                      <div className="overflow-x-auto">
                        <table className="min-w-full text-xs">
                          <thead>
                            <tr className="text-gray-500">
                              <th className="pr-6 py-1 text-left font-semibold">Month</th>
                              <th className="px-4 py-1 text-right font-semibold">Total</th>
                              <th className="px-4 py-1 text-right font-semibold">Max</th>
                              <th className="px-4 py-1 text-right font-semibold">Avg</th>
                              <th className="px-4 py-1 text-right font-semibold">Txs</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-indigo-100">
                            {merchant.months.map((mo) => (
                              <tr key={mo.month} className="text-gray-700">
                                <td className="pr-6 py-1.5 font-medium tabular-nums">{mo.month}</td>
                                <td className="px-4 py-1.5 text-right tabular-nums font-semibold">€{fmt(mo.total)}</td>
                                <td className="px-4 py-1.5 text-right tabular-nums">€{fmt(mo.maxAmount)}</td>
                                <td className="px-4 py-1.5 text-right tabular-nums">€{fmt(mo.avgAmount)}</td>
                                <td className="px-4 py-1.5 text-right tabular-nums text-gray-500">{mo.count}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </td>
                  </tr>
                )}
              </>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

// ─── Section wrapper ──────────────────────────────────────────────────────────

function Section({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <section className="space-y-4">
      <div>
        <h3 className="text-base font-semibold text-gray-900">{title}</h3>
        {subtitle && <p className="mt-0.5 text-sm text-gray-500">{subtitle}</p>}
      </div>
      {children}
    </section>
  );
}

// ─── ReportsView ──────────────────────────────────────────────────────────────

interface ReportState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

function useReport<T>(fetcher: () => Promise<T>): ReportState<T> {
  const [state, setState] = useState<ReportState<T>>({
    data: null,
    loading: true,
    error: null,
  });

  const load = useCallback(async () => {
    setState({ data: null, loading: true, error: null });
    try {
      const data = await fetcher();
      setState({ data, loading: false, error: null });
    } catch (err) {
      setState({
        data: null,
        loading: false,
        error: err instanceof Error ? err.message : "Unexpected error",
      });
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => { load(); }, [load]);

  return state;
}

export function ReportsView() {
  const primary = useReport(getSpendByPrimaryTag);
  const secondary = useReport(getSpendBySecondaryTag);
  const merchants = useReport(getMerchantsByMonth);

  return (
    <div className="space-y-10">
      <div>
        <h2 className="text-xl font-semibold text-gray-900">Reports</h2>
        <p className="mt-0.5 text-sm text-gray-500">
          Aggregated spending analysis across your transactions.
        </p>
      </div>

      {/* Spending by primary tag */}
      <Section
        title="Spending by primary tag"
        subtitle="Total debit amount for each primary tag (only tagged merchants)."
      >
        <div className="rounded-xl border border-gray-200 bg-white shadow-sm px-5 py-5">
          {primary.loading ? (
            <Skeleton />
          ) : primary.error ? (
            <SectionError message={primary.error} />
          ) : (
            <TagSpendList items={primary.data?.items ?? []} />
          )}
        </div>
      </Section>

      {/* Spending by secondary tag */}
      <Section
        title="Spending by secondary tag"
        subtitle="Total debit amount broken down by secondary classification."
      >
        <div className="rounded-xl border border-gray-200 bg-white shadow-sm px-5 py-5">
          {secondary.loading ? (
            <Skeleton />
          ) : secondary.error ? (
            <SectionError message={secondary.error} />
          ) : (
            <TagSpendList items={secondary.data?.items ?? []} />
          )}
        </div>
      </Section>

      {/* Top merchants by month */}
      <Section
        title="Merchants by month"
        subtitle="Monthly spend per merchant, sorted by total. Click a row to see max and average per month."
      >
        {merchants.loading ? (
          <Skeleton />
        ) : merchants.error ? (
          <SectionError message={merchants.error} />
        ) : (
          <MerchantsByMonthTable items={merchants.data?.items ?? []} />
        )}
      </Section>
    </div>
  );
}
