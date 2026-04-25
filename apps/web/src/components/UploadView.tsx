import { useCallback, useRef, useState } from "react";

type Phase = "idle" | "processing" | "done";

const PROCESSING_MS = 5000;

export function UploadView() {
  const [phase, setPhase] = useState<Phase>("idle");
  const [fileName, setFileName] = useState<string | null>(null);
  const [progress, setProgress] = useState(0);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const startProcessing = useCallback((file: File) => {
    setFileName(file.name);
    setPhase("processing");
    setProgress(0);

    const start = Date.now();
    timerRef.current = setInterval(() => {
      const elapsed = Date.now() - start;
      const pct = Math.min((elapsed / PROCESSING_MS) * 100, 100);
      setProgress(pct);
      if (pct >= 100) {
        clearInterval(timerRef.current!);
        setPhase("done");
      }
    }, 50);
  }, []);

  const handleFile = useCallback(
    (file: File | undefined) => {
      if (!file) return;
      if (phase !== "idle" && phase !== "done") return;
      setPhase("idle");
      setProgress(0);
      // kick off on next tick so reset renders first
      setTimeout(() => startProcessing(file), 10);
    },
    [phase, startProcessing]
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
    clearInterval(timerRef.current!);
    setPhase("idle");
    setFileName(null);
    setProgress(0);
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

        {/* Drop zone */}
        {(phase === "idle" || phase === "done") && (
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
              <p className="mt-1 text-xs text-gray-400">Supported: Piraeus Bank CSV/PDF statements</p>
            </div>
            {phase === "done" && fileName && (
              <div className="absolute top-3 right-3 flex items-center gap-1.5 rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-700">
                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                </svg>
                Imported
              </div>
            )}
          </div>
        )}

        {/* Processing state */}
        {phase === "processing" && (
          <div className="rounded-2xl border border-gray-200 bg-white shadow-sm px-8 py-10 flex flex-col items-center gap-6">
            {/* Spinner */}
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

            {/* Steps */}
            <ProcessingSteps progress={progress} />

            {/* Progress bar */}
            <div className="w-full space-y-1.5">
              <div className="h-2 w-full overflow-hidden rounded-full bg-gray-100">
                <div
                  className="h-full rounded-full bg-indigo-500 transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <p className="text-right text-xs text-gray-400 tabular-nums">{Math.round(progress)}%</p>
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
              <p className="mt-1 text-xs text-emerald-700">Transactions are now available in the Transactions tab.</p>
            </div>
            <button
              onClick={reset}
              className="shrink-0 text-xs text-emerald-600 hover:text-emerald-800 font-medium underline underline-offset-2"
            >
              Upload another
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

const STEPS = [
  { at: 0,  label: "Reading file…" },
  { at: 20, label: "Parsing PDF structure…" },
  { at: 45, label: "Extracting transactions…" },
  { at: 70, label: "Normalising records…" },
  { at: 88, label: "Saving to database…" },
];

function ProcessingSteps({ progress }: { progress: number }) {
  return (
    <ol className="w-full space-y-1.5">
      {STEPS.map((step, i) => {
        const done = progress >= (STEPS[i + 1]?.at ?? 101);
        const active = progress >= step.at && !done;
        return (
          <li key={step.label} className={`flex items-center gap-2.5 text-xs transition-colors ${done ? "text-emerald-600" : active ? "text-indigo-600 font-medium" : "text-gray-300"}`}>
            {done ? (
              <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
            ) : active ? (
              <svg className="h-4 w-4 shrink-0 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4l3-3-3-3V4a8 8 0 110 8z" />
              </svg>
            ) : (
              <span className="h-4 w-4 shrink-0 rounded-full border-2 border-current" />
            )}
            {step.label}
          </li>
        );
      })}
    </ol>
  );
}
