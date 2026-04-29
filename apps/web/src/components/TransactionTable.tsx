import { type Transaction } from "../api/transactions";
import { penceToDisplay } from "../lib/money";

interface Props {
  transactions: Transaction[];
}

const PAGE_LIMIT = 50;

function amountClass(direction: string): string {
  return direction === "credit"
    ? "text-emerald-600 font-medium"
    : "text-red-500 font-medium";
}

function formatAmount(amount: number, direction: string): string {
  const prefix = direction === "credit" ? "+" : "-";
  return `${prefix}£${penceToDisplay(amount)}`;
}

export function TransactionTable({ transactions }: Props) {
  if (transactions.length === 0) {
    return (
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
            d="M9 17v-2a4 4 0 014-4h4M3 7h18M3 12h6"
          />
        </svg>
        <p className="text-sm">No transactions found</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">
              Date
            </th>
            <th className="px-4 py-3 text-left font-semibold text-gray-500 uppercase tracking-wide text-xs">
              Description
            </th>
            <th className="px-4 py-3 text-right font-semibold text-gray-500 uppercase tracking-wide text-xs">
              Amount
            </th>
            <th className="px-4 py-3 text-right font-semibold text-gray-500 uppercase tracking-wide text-xs">
              Balance
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {transactions.map((t) => (
            <tr key={t.id} className="hover:bg-gray-50 transition-colors">
              <td className="px-4 py-3 whitespace-nowrap text-gray-500 tabular-nums">
                {t.date}
              </td>
              <td className="px-4 py-3 text-gray-800 max-w-xs truncate">
                {t.transactionReference || t.merchantIdentifier || "—"}
              </td>
              <td
                className={`px-4 py-3 whitespace-nowrap text-right tabular-nums ${amountClass(t.direction)}`}
              >
                {formatAmount(t.amount, t.direction)}
              </td>
              <td className="px-4 py-3 whitespace-nowrap text-right text-gray-500 tabular-nums">
                £{penceToDisplay(t.balanceAfter)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export { PAGE_LIMIT };
