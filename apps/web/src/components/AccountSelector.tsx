import { type Account } from "../api/accounts";

interface Props {
  accounts: Account[];
  selectedIds: string[];
  onChange: (ids: string[]) => void;
}

export function AccountSelector({ accounts, selectedIds, onChange }: Props) {
  if (accounts.length <= 1) return null;

  const allSelected = selectedIds.length === 0;

  function toggle(id: string) {
    if (selectedIds.includes(id)) {
      const next = selectedIds.filter((s) => s !== id);
      onChange(next);
    } else {
      onChange([...selectedIds, id]);
    }
  }

  return (
    <div className="flex items-center gap-2 flex-wrap mb-5">
      <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">Account</span>
      <button
        onClick={() => onChange([])}
        className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
          allSelected
            ? "bg-indigo-600 text-white"
            : "bg-gray-100 text-gray-600 hover:bg-gray-200"
        }`}
      >
        All
      </button>
      {accounts.map((a) => (
        <button
          key={a.id}
          onClick={() => toggle(a.id)}
          className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
            selectedIds.includes(a.id)
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-600 hover:bg-gray-200"
          }`}
        >
          {a.bankName} {a.accountNumber}
        </button>
      ))}
    </div>
  );
}
