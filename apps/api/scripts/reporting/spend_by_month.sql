-- Month-by-month debits and credits (respects transaction_extra.in_report)
SELECT
    date_trunc('month', t.date)::date AS spend_month,
    SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END) AS debits,
    SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END) AS credits
FROM transactions t
LEFT JOIN transaction_extra te ON te.transaction_id = t.id
WHERE COALESCE(te.in_report, true)
GROUP BY 1
ORDER BY 1;
