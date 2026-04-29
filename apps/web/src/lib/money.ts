/**
 * Converts an integer pence value to a human-readable pounds string.
 * e.g. 1223 → "12.23"  |  100000 → "1,000.00"  |  0 → "0.00"
 */
export function penceToDisplay(pence: number): string {
  return (pence / 100).toLocaleString("en-GB", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

/**
 * Parses a string that contains an integer pence value and formats it as pounds.
 * e.g. "1223" → "12.23"  |  "100000" → "1,000.00"
 */
export function penceToPounds(penceStr: string): string {
  const n = parseInt(penceStr, 10);
  if (isNaN(n)) return penceStr;
  return penceToDisplay(n);
}

/**
 * Compact format for tight spaces: rounds to nearest pound with a k-suffix for
 * values >= £1,000.
 * e.g. 122300 → "£1.2k"  |  1223 → "£12"
 */
export function penceCompact(penceStr: string): string {
  const n = parseInt(penceStr, 10);
  if (isNaN(n)) return "—";
  const pounds = n / 100;
  if (pounds >= 1000) return `£${(pounds / 1000).toFixed(1)}k`;
  return `£${pounds.toFixed(0)}`;
}
