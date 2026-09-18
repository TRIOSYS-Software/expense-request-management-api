package repositories

type SyncCounts struct {
	Upserted      int64
	Deleted       int64
	Renamed       map[string]string
	RenameSkipped []string
}
