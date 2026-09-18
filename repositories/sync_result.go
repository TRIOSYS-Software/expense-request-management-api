package repositories

// SyncCounts is what a transactional reconcile-and-upsert reports back to
// the caller: how many rows were affected by each step, plus the keys that
// disappeared upstream but are still referenced locally and so were kept.
type SyncCounts struct {
	Upserted int64
	Deleted  int64
	Retained []string
}
