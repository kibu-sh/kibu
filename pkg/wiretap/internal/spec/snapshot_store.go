package spec

type SnapshotStore interface {
	Read(ref *SnapshotRef) (snapshot *Snapshot, err error)
	Write(snapshot *Snapshot) (ref *SnapshotRef, err error)
}
