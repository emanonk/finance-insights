import { useEffect, useState, useCallback } from "react";
import { Routes, Route, NavLink, Navigate, useNavigate } from "react-router-dom";
import {
  listTransactions,
  type Transaction,
} from "./api/transactions";
import {
  getTopIdentifiers,
  type Merchant,
} from "./api/merchants";
import { listAccounts, type Account } from "./api/accounts";
import { TransactionTable, PAGE_LIMIT } from "./components/TransactionTable";
import { AccountSelector } from "./components/AccountSelector";
import { UploadView } from "./components/UploadView";
import { CalendarView } from "./components/CalendarView";
import { MerchantsView } from "./components/MerchantsView";
import { ReportsView } from "./components/ReportsView";

const TABS = [
  { path: "/upload", label: "Upload" },
  { path: "/transactions", label: "Transactions" },
  { path: "/calendar", label: "Calendar" },
  { path: "/merchants", label: "Merchants" },
  { path: "/reports", label: "Reports" },
];

// ─── Transactions page ────────────────────────────────────────────────────────

interface TxState {
  transactions: Transaction[];
  total: number;
  loading: boolean;
  error: string | null;
}

interface TransactionsPageProps {
  accounts: Account[];
  selectedAccountIds: string[];
  onAccountsChange: (ids: string[]) => void;
}

function TransactionsPage({ accounts, selectedAccountIds, onAccountsChange }: TransactionsPageProps) {
  const [page, setPage] = useState(0);
  const [state, setState] = useState<TxState>({
    transactions: [],
    total: 0,
    loading: true,
    error: null,
  });
  const [merchants, setMerchants] = useState<Map<string, Merchant>>(new Map());

  // Reset to page 0 when account filter changes
  useEffect(() => { setPage(0); }, [selectedAccountIds.join(",")]);

  const loadTx = useCallback(async (offset: number, accountIds: string[]) => {
    setState((s) => ({ ...s, loading: true, error: null }));
    try {
      const data = await listTransactions(PAGE_LIMIT, offset, accountIds.length > 0 ? accountIds : undefined);
      setState({ transactions: data.items, total: data.total, loading: false, error: null });
    } catch (err) {
      setState((s) => ({ ...s, loading: false, error: err instanceof Error ? err.message : "Unexpected error" }));
    }
  }, []);

  useEffect(() => {
    loadTx(page * PAGE_LIMIT, selectedAccountIds);
  }, [page, loadTx, selectedAccountIds.join(",")]);

  // Load merchant map once (high limit to cover all identifiers)
  useEffect(() => {
    getTopIdentifiers(2000).then((data) => {
      const map = new Map<string, Merchant>();
      for (const item of data.items) {
        if (item.merchant) map.set(item.identifier, item.merchant);
      }
      setMerchants(map);
    }).catch(() => {/* non-fatal */});
  }, []);

  function handleMerchantSaved(merchant: Merchant) {
    setMerchants((prev) => {
      const next = new Map(prev);
      next.set(merchant.identifierName, merchant);
      return next;
    });
  }

  const totalPages = Math.ceil(state.total / PAGE_LIMIT);
  const firstItem = state.total === 0 ? 0 : page * PAGE_LIMIT + 1;
  const lastItem = Math.min((page + 1) * PAGE_LIMIT, state.total);

  return (
    <>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Transactions</h2>
          {!state.loading && !state.error && (
            <p className="mt-0.5 text-sm text-gray-500">
              {state.total === 0
                ? "No transactions"
                : `${firstItem}–${lastItem} of ${state.total}`}
            </p>
          )}
        </div>
      </div>

      <AccountSelector accounts={accounts} selectedIds={selectedAccountIds} onChange={onAccountsChange} />

      {state.error && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {state.error}
        </div>
      )}

      {state.loading ? (
        <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex gap-4 border-b border-gray-100 px-4 py-3 last:border-0">
              <div className="h-4 w-20 animate-pulse rounded bg-gray-100" />
              <div className="h-4 flex-1 animate-pulse rounded bg-gray-100" />
              <div className="h-4 w-24 animate-pulse rounded bg-gray-100" />
              <div className="h-4 w-20 animate-pulse rounded bg-gray-100" />
              <div className="h-4 w-24 animate-pulse rounded bg-gray-100" />
            </div>
          ))}
        </div>
      ) : (
        <TransactionTable
          transactions={state.transactions}
          merchants={merchants}
          onMerchantSaved={handleMerchantSaved}
        />
      )}

      {!state.loading && totalPages > 1 && (
        <div className="mt-6 flex items-center justify-between">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
          >
            ← Previous
          </button>
          <span className="text-sm text-gray-500">Page {page + 1} of {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
          >
            Next →
          </button>
        </div>
      )}
    </>
  );
}

// ─── App shell ────────────────────────────────────────────────────────────────

function App() {
  const navigate = useNavigate();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccountIds, setSelectedAccountIds] = useState<string[]>([]);

  useEffect(() => {
    listAccounts()
      .then((data) => setAccounts(data.items))
      .catch(() => {/* non-fatal — selector just won't appear */});
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white">
        <div className="mx-auto max-w-6xl px-6 py-4 flex items-center gap-6">
          <div className="flex items-center gap-3 shrink-0">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h1 className="text-lg font-semibold text-gray-900">Finance Insights</h1>
          </div>

          {/* Nav tabs */}
          <nav className="flex gap-1">
            {TABS.map((tab) => (
              <NavLink
                key={tab.path}
                to={tab.path}
                className={({ isActive }) =>
                  `rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-indigo-50 text-indigo-700"
                      : "text-gray-500 hover:text-gray-800 hover:bg-gray-100"
                  }`
                }
              >
                {tab.label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>

      {/* Main */}
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Routes>
          <Route path="/" element={<Navigate to="/upload" replace />} />
          <Route path="/upload" element={<UploadView onImported={() => navigate("/transactions")} />} />
          <Route
            path="/transactions"
            element={
              <TransactionsPage
                accounts={accounts}
                selectedAccountIds={selectedAccountIds}
                onAccountsChange={setSelectedAccountIds}
              />
            }
          />
          <Route
            path="/calendar"
            element={
              <CalendarView
                accounts={accounts}
                selectedAccountIds={selectedAccountIds}
                onAccountsChange={setSelectedAccountIds}
              />
            }
          />
          <Route path="/merchants" element={<MerchantsView />} />
          <Route
            path="/reports"
            element={
              <ReportsView
                accounts={accounts}
                selectedAccountIds={selectedAccountIds}
                onAccountsChange={setSelectedAccountIds}
              />
            }
          />
          <Route path="*" element={<Navigate to="/upload" replace />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
