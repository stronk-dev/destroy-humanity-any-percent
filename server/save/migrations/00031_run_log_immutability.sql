-- +goose Up
-- A run log is forensic replay evidence. Corrections append new evidence;
-- neither active rows nor their replay inputs may be rewritten or removed.
CREATE TRIGGER run_log_immutable
BEFORE UPDATE OR DELETE ON run_log
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER run_log_immutable ON run_log;
