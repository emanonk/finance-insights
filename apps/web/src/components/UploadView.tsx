import { useCallback, useRef, useState } from "react";
import { uploadStatement } from "../api/statements";

interface Bank {
  id: string;
  label: string;
  supported: boolean;
}

const BANKS: Bank[] = [
  { id: "piraeus",    label: "Piraeus Bank", supported: true  },
  { id: "alpha",      label: "Alpha Bank",   supported: false },
  { id: "nbg",        label: "NBG",          supported: false },
  { id: "eurobank",   label: "Eurobank",     supported: false },
];

type Phase = "select-bank" | "idle" | "processing" | "done" | "error";

interface UploadViewProps {
  onImported?: () => void;
}

export function UploadView({ onImported }: UploadViewProps) {
  const [selectedBank, setSelectedBank] = useState<Bank | null>(null);
  const [phase, setPhase] = useState<Phase>("select-bank");
  const [fileName, setFileName] = useState<string | null>(null);
  const [transactionCount, setTransactionCount] = useState<number | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleBankSelect = useCallback((bank: Bank) => {
    if (!bank.supported) return;
    setSelectedBank(bank);
    setPhase("idle");
  }, []);

  const handleFile = useCallback(
    async (file: File | undefined) => {
      if (!file || !selectedBank) return;
      if (phase === "processing") return;

      setFileName(file.name);
      setPhase("processing");
      setErrorMsg(null);
      setTransactionCount(null);

      try {
        const result = await uploadStatement(selectedBank.id, file);
        setTransactionCount(result.transactionCount);
        setPhase("done");
      } catch (err) {
        setErrorMsg(err instanceof Error ? err.message : "Upload failed");
        setPhase("error");
      }
    },
    [phase, selectedBank]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      handleFile(e.dataTransfer.files[0]);
    },
    [handleFile]
  );

  function reset() {
    setPhase("idle");
    setFileName(null);
    setTransactionCount(null);
    setErrorMsg(null);
  }

  function resetAll() {
    setSelectedBank(null);
    setPhase("select-bank");
    setFileName(null);
    setTransactionCount(null);
    setErrorMsg(null);
  }

  return (
    <div className="flex flex-col items-center justify-start pt-8 pb-16">
      <div className="w-full max-w-xl space-y-6">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Upload Statement</h2>
          <p className="mt-0.5 text-sm text-gray-500">
            Upload a bank statement PDF and we'll parse and import your transactions automatically.
          </p>
        </div>

        {/* Bank selection */}
        {phase === "select-bank" && (
          <div className="space-y-3">
            <p className="text-sm font-medium text-gray-700">Select your bank</p>
            <div className="grid grid-cols-2 gap-3">
              {BANKS.map((bank) => (
                <button
                  key={bank.id}
                  onClick={() => handleBankSelect(bank)}
                  disabled={!bank.supported}
                  className={`relative flex flex-col items-start gap-1 rounded-xl border-2 px-4 py-4 text-left transition-colors ${
                    bank.supported
                      ? "border-gray-200 bg-white hover:border-indigo-400 hover:bg-indigo-50/40 cursor-pointer"
                      : "border-gray-100 bg-gray-50 cursor-not-allowed opacity-60"
                  }`}
                >
                  <span className="text-sm font-semibold text-gray-800">{bank.label}</span>
                  {!bank.supported && (
                    <span className="text-xs text-gray-400">Coming soon</span>
                  )}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Upload zone — shown after bank is selected */}
        {(phase === "idle" || phase === "done" || phase === "error") && selectedBank && (
          <>
            {/* Selected bank indicator */}
            <div className="flex items-center justify-between rounded-lg border border-indigo-200 bg-indigo-50 px-4 py-2.5">
              <span className="text-sm font-medium text-indigo-800">{selectedBank.label}</span>
              <button
                onClick={resetAll}
                className="text-xs text-indigo-500 hover:text-indigo-700 font-medium underline underline-offset-2"
              >
                Change bank
              </button>
            </div>

            <div
              onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
              onDragLeave={() => setDragOver(false)}
              onDrop={handleDrop}
              onClick={() => inputRef.current?.click()}
              className={`relative flex flex-col items-center justify-center gap-4 rounded-2xl border-2 border-dashed px-8 py-16 cursor-pointer transition-colors ${
                dragOver
                  ? "border-indigo-400 bg-indigo-50"
                  : "border-gray-300 bg-white hover:border-indigo-400 hover:bg-indigo-50/40"
              }`}
            >
              <input
                ref={inputRef}
                type="file"
                accept=".pdf"
                className="sr-only"
                onChange={(e) => handleFile(e.target.files?.[0])}
              />
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-indigo-100">
                <svg className="h-7 w-7 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
                </svg>
              </div>
              <div className="text-center">
                <p className="text-sm font-semibold text-gray-700">
                  Drop your PDF here, or{" "}
                  <span className="text-indigo-600">browse</span>
                </p>
                <p className="mt-1 text-xs text-gray-400">PDF statements only · max 25 MB</p>
              </div>
              {phase === "done" && (
                <div className="absolute top-3 right-3 flex items-center gap-1.5 rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-700">
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                  </svg>
                  Imported
                </div>
              )}
            </div>
          </>
        )}

        {/* Processing state */}
        {phase === "processing" && (
          <div className="rounded-2xl border border-gray-200 bg-white shadow-sm px-8 py-10 flex flex-col items-center gap-6">
            <div className="relative flex h-20 w-20 items-center justify-center">
              <svg className="absolute inset-0 h-20 w-20 animate-spin text-indigo-200" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
              </svg>
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-50">
                <svg className="h-6 w-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </svg>
              </div>
            </div>
            <div className="text-center space-y-1">
              <p className="text-sm font-semibold text-gray-800">Processing statement…</p>
              <p className="text-xs text-gray-500 font-mono truncate max-w-xs">{fileName}</p>
            </div>
          </div>
        )}

        {/* Success state */}
        {phase === "done" && fileName && (
          <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-6 py-5 flex items-start gap-4">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-emerald-100">
              <svg className="h-5 w-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-emerald-800">Statement imported successfully</p>
              <p className="mt-0.5 text-xs text-emerald-600 font-mono truncate">{fileName}</p>
              {transactionCount !== null && (
                <p className="mt-1 text-xs text-emerald-700">
                  {transactionCount} transaction{transactionCount !== 1 ? "s" : ""} imported.{" "}
                  {onImported && (
                    <button
                      onClick={onImported}
                      className="underline underline-offset-2 hover:text-emerald-900"
                    >
                      View in Transactions
                    </button>
                  )}
                </p>
              )}
            </div>
            <button
              onClick={reset}
              className="shrink-0 text-xs text-emerald-600 hover:text-emerald-800 font-medium underline underline-offset-2"
            >
              Upload another
            </button>
          </div>
        )}

        {/* Error state */}
        {phase === "error" && errorMsg && (
          <div className="rounded-2xl border border-red-200 bg-red-50 px-6 py-5 flex items-start gap-4">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-red-100">
              <svg className="h-5 w-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
              </svg>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-red-800">Upload failed</p>
              <p className="mt-0.5 text-xs text-red-600">{errorMsg}</p>
            </div>
            <button
              onClick={reset}
              className="shrink-0 text-xs text-red-600 hover:text-red-800 font-medium underline underline-offset-2"
            >
              Try again
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
