-- +goose Up

-- This migration backfills ledger entries so that for every account:
-- accounts.balance == SUM(ledger.amount)
--
-- It introduces a system "equity / opening balances" account that acts as the
-- counterparty for one-time opening-balance postings (standard double-entry practice).
-- The equity account balance may become negative, so we adjust the accounts balance
-- check constraint accordingly (negative balances are allowed ONLY for this system user).

UPDATE users
SET
  email = 'equity@system.local',
  first_name = 'System',
  last_name = 'Equity',
  updated_at = NOW()
WHERE id = '00000000-0000-0000-0000-000000000002'::uuid;

-- Create equity user (system, cannot login).
INSERT INTO users (id, email, password, first_name, last_name, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'equity@system.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'System',
    'Equity',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- Drop existing balance check constraint(s) that reference "balance".
-- +goose StatementBegin
DO $$
DECLARE r record;
BEGIN
  FOR r IN
    SELECT conname
    FROM pg_constraint
    WHERE conrelid = 'accounts'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) ILIKE '%balance%'
  LOOP
    EXECUTE format('ALTER TABLE accounts DROP CONSTRAINT %I', r.conname);
  END LOOP;
END $$;
-- +goose StatementEnd

-- Re-add balance check allowing negative only for equity user.
ALTER TABLE accounts
  ADD CONSTRAINT accounts_balance_check
  CHECK (balance >= 0 OR user_id = '00000000-0000-0000-0000-000000000002'::uuid);

-- Create equity accounts (start at 0; will be updated to match ledger after backfill).
INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000002', 'USD', 0.00, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000022', '00000000-0000-0000-0000-000000000002', 'EUR', 0.00, NOW(), NOW())
ON CONFLICT (user_id, currency) DO NOTHING;

-- Insert reconciliation transfers and corresponding ledger entries.
WITH
equity_usd AS (
  SELECT id FROM accounts WHERE user_id = '00000000-0000-0000-0000-000000000002'::uuid AND currency = 'USD' LIMIT 1
),
equity_eur AS (
  SELECT id FROM accounts WHERE user_id = '00000000-0000-0000-0000-000000000002'::uuid AND currency = 'EUR' LIMIT 1
),
ledger_sums AS (
  SELECT account_id, COALESCE(SUM(amount), 0.00)::numeric(15,2) AS ledger_sum
  FROM ledger
  GROUP BY account_id
),
diffs AS (
  SELECT
    a.id AS account_id,
    a.currency,
    (a.balance - COALESCE(ls.ledger_sum, 0.00))::numeric(15,2) AS diff_amount
  FROM accounts a
  LEFT JOIN ledger_sums ls ON ls.account_id = a.id
  WHERE a.user_id <> '00000000-0000-0000-0000-000000000002'::uuid
    AND (a.balance - COALESCE(ls.ledger_sum, 0.00)) <> 0.00
),
ins AS (
  INSERT INTO transactions (id, type, from_account_id, to_account_id, amount, currency, exchange_rate, converted_amount, description, created_at)
  SELECT
    gen_random_uuid(),
    'transfer',
    CASE
      WHEN d.diff_amount > 0 THEN (CASE d.currency WHEN 'USD' THEN (SELECT id FROM equity_usd) ELSE (SELECT id FROM equity_eur) END)
      ELSE d.account_id
    END,
    CASE
      WHEN d.diff_amount > 0 THEN d.account_id
      ELSE (CASE d.currency WHEN 'USD' THEN (SELECT id FROM equity_usd) ELSE (SELECT id FROM equity_eur) END)
    END,
    ABS(d.diff_amount),
    d.currency,
    NULL,
    NULL,
    'opening_balance_reconciliation',
    NOW()
  FROM diffs d
  RETURNING id, from_account_id, to_account_id, amount, created_at
)
INSERT INTO ledger (id, transaction_id, account_id, amount, created_at)
SELECT gen_random_uuid(), id, from_account_id, -amount, created_at FROM ins
UNION ALL
SELECT gen_random_uuid(), id, to_account_id, amount, created_at FROM ins;

-- Update equity cached balances to match ledger.
UPDATE accounts a
SET balance = COALESCE(ls.sum_amount, 0.00)
FROM (
  SELECT account_id, COALESCE(SUM(amount), 0.00)::numeric(15,2) AS sum_amount
  FROM ledger
  GROUP BY account_id
) ls
WHERE a.user_id = '00000000-0000-0000-0000-000000000002'::uuid
  AND a.id = ls.account_id;

-- +goose Down

-- Best-effort rollback: remove reconciliation transactions (ledger cascades).
DELETE FROM transactions WHERE description = 'opening_balance_reconciliation';

-- Remove equity accounts and user (may fail if FK references exist).
DELETE FROM accounts WHERE user_id = '00000000-0000-0000-0000-000000000002'::uuid;
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000002'::uuid;

