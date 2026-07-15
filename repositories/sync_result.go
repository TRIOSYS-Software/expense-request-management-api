package repositories

// SyncCounts is what a transactional reconcile-and-upsert reports back to
// the caller: how many rows were affected by each step.
type SyncCounts struct {
	Upserted int64
	Deleted  int64
}
