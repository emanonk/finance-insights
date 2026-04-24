-- Recurring-like debits: same merchant + same rounded amount, full-year threshold (>=10 per calendar year)
SELECT
    EXTRACT(YEAR FROM t.date)::int AS calendar_year,
    t.merchant_identifier,
    round(t.amount::numeric, 2) AS charge_amount,
    COUNT(*) AS occurrences,
    SUM(t.amount) AS total_debited,
    MIN(t.date) AS first_date,
    MAX(t.date) AS last_date
FROM transactions t
LEFT JOIN transaction_extra te ON te.transaction_id = t.id
WHERE t.direction = 'debit'
  AND t.merchant_identifier IS NOT NULL
  AND COALESCE(te.in_report, true)
GROUP BY calendar_year, t.merchant_identifier, round(t.amount::numeric, 2)
HAVING COUNT(*) >= 10
ORDER BY calendar_year, total_debited DESC;
