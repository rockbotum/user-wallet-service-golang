-- Remove seed data
DELETE FROM transaction_types WHERE id IN (1, 2, 3, 4);
DELETE FROM roles WHERE id IN (1, 2);