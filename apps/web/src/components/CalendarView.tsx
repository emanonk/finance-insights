import { useEffect, useState } from "react";
import { getDailySpend } from "../api/reports";

const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

function fmt(value: string): string {
  const n = parseFloat(value);
  if (isNaN(n)) return "—";
  if (n >= 1000) return `€${(n / 1000).toFixed(1)}k`;
  return `€${n.toFixed(0)}`;
}

function fmtFull(value: string): string {
  const n = parseFloat(value);
  if (isNaN(n)) return "—";
  return n.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

// Returns 0=Mon … 6=Sun offset for day 1 of a month
function monthStartOffset(year: number, month: number): number {
  const d = new Date(year, month - 1, 1).getDay(); // 0=Sun
  return (d + 6) % 7; // shift so Mon=0
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month, 0).getDate();
}

// Heat-map: returns a Tailwind bg class based on intensity 0..1
function heatClass(intensity: number): string {
  if (intensity === 0) return "";
  if (intensity < 0.15) return "bg-indigo-50";
  if (intensity < 0.3) return "bg-indigo-100";
  if (intensity < 0.5) return "bg-indigo-200";
  if (intensity < 0.7) return "bg-indigo-300";
  if (intensity < 0.85) return "bg-indigo-400";
  return "bg-indigo-500";
}

function textClass(intensity: number): string {
  return intensity >= 0.7 ? "text-white" : "text-gray-700";
}

function amountTextClass(intensity: number): string {
  return intensity >= 0.7 ? "text-indigo-100" : "text-indigo-600";
}

export function CalendarView() {
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1); // 1-based
  const [spendMap, setSpendMap] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tooltip, setTooltip] = useState<{ day: number; total: string } | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getDailySpend()
      .then((res) => {
        const map: Record<string, string> = {};
        for (const d of res.items) map[d.date] = d.total;
        setSpendMap(map);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Unexpected error"))
      .finally(() => setLoading(false));
  }, []);

  function prevMonth() {
    if (month === 1) { setYear(y => y - 1); setMonth(12); }
    else setMonth(m => m - 1);
  }
  function nextMonth() {
    if (month === 12) { setYear(y => y + 1); setMonth(1); }
    else setMonth(m => m + 1);
  }

  const label = new Date(year, month - 1, 1).toLocaleString("en-US", { month: "long", year: "numeric" });
  const offset = monthStartOffset(year, month);
  const days = daysInMonth(year, month);

  // Collect spend values for this month to compute heat-map scale.
  const monthValues: number[] = [];
  for (let d = 1; d <= days; d++) {
    const key = `${year}-${String(month).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
    const v = spendMap[key];
    if (v) monthValues.push(parseFloat(v) || 0);
  }
  const maxVal = monthValues.length ? Math.max(...monthValues) : 1;

  // Monthly total
  const monthTotal = monthValues.reduce((a, b) => a + b, 0);

  // Build grid cells: offset blanks + day cells.
  const cells: Array<{ blank: true } | { day: number; spend: string | null }> = [
    ...Array.from({ length: offset }, () => ({ blank: true as const })),
    ...Array.from({ length: days }, (_, i) => {
      const d = i + 1;
      const key = `${year}-${String(month).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
      return { day: d, spend: spendMap[key] ?? null };
    }),
  ];

  const todayKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Spending Calendar</h2>
          <p className="mt-0.5 text-sm text-gray-500">Daily debit totals — darker cells mean higher spend.</p>
        </div>
        {!loading && !error && monthTotal > 0 && (
          <div className="text-right">
            <p className="text-xs text-gray-400 uppercase tracking-wide font-semibold">Month total</p>
            <p className="text-lg font-bold text-indigo-600">€{fmtFull(String(monthTotal))}</p>
          </div>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
      )}

      <div className="rounded-2xl border border-gray-200 bg-white shadow-sm overflow-hidden">
        {/* Month nav header */}
        <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
          <button
            onClick={prevMonth}
            className="rounded-lg p-1.5 text-gray-400 hover:text-gray-700 hover:bg-gray-100 transition-colors"
            aria-label="Previous month"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <span className="text-sm font-semibold text-gray-900">{label}</span>
          <button
            onClick={nextMonth}
            className="rounded-lg p-1.5 text-gray-400 hover:text-gray-700 hover:bg-gray-100 transition-colors"
            aria-label="Next month"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>

        {/* Day-of-week headers */}
        <div className="grid grid-cols-7 border-b border-gray-100">
          {DAYS.map((d) => (
            <div key={d} className="py-2 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">
              {d}
            </div>
          ))}
        </div>

        {/* Calendar grid */}
        {loading ? (
          <div className="grid grid-cols-7 gap-px bg-gray-100 p-px">
            {Array.from({ length: 35 }).map((_, i) => (
              <div key={i} className="h-20 animate-pulse bg-white" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-7 gap-px bg-gray-100 p-px">
            {cells.map((cell, i) => {
              if ("blank" in cell) {
                return <div key={`b${i}`} className="h-20 bg-gray-50" />;
              }
              const { day, spend } = cell;
              const dayKey = `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
              const isToday = dayKey === todayKey;
              const val = spend ? parseFloat(spend) || 0 : 0;
              const intensity = maxVal > 0 ? val / maxVal : 0;

              return (
                <div
                  key={day}
                  className={`relative h-20 p-2 flex flex-col justify-between transition-all cursor-default ${heatClass(intensity)} ${isToday ? "ring-2 ring-inset ring-indigo-500" : ""}`}
                  onMouseEnter={() => spend ? setTooltip({ day, total: spend }) : undefined}
                  onMouseLeave={() => setTooltip(null)}
                >
                  <span className={`text-xs font-semibold leading-none ${isToday ? "text-indigo-600" : textClass(intensity)}`}>
                    {day}
                  </span>
                  {spend && (
                    <span className={`text-xs font-bold tabular-nums leading-none ${amountTextClass(intensity)}`}>
                      {fmt(spend)}
                    </span>
                  )}
                  {/* Tooltip */}
                  {tooltip?.day === day && spend && (
                    <div className="absolute bottom-full left-1/2 z-20 mb-1 -translate-x-1/2 whitespace-nowrap rounded-lg bg-gray-900 px-2.5 py-1.5 text-xs text-white shadow-lg pointer-events-none">
                      €{fmtFull(spend)}
                      <div className="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900" />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}

        {/* Legend */}
        <div className="flex items-center justify-end gap-2 border-t border-gray-100 px-5 py-3">
          <span className="text-xs text-gray-400">Less</span>
          {[0, 0.2, 0.4, 0.6, 0.8, 1].map((v) => (
            <div key={v} className={`h-4 w-4 rounded-sm ${v === 0 ? "bg-gray-100" : heatClass(v)}`} />
          ))}
          <span className="text-xs text-gray-400">More</span>
        </div>
      </div>
    </div>
  );
}
