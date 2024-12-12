package sqlx

type Option string

func (o Option) String() string {
	return string(o)
}

type Cache = Option

const (
	CACHE_Shared  Cache = "cache=shared"
	CACHE_Private Cache = "cache=private"
)

type Synchronous = Option

const (
	SYNC_OFF    Synchronous = "_synchronous=OFF"
	SYNC_NORMAL Synchronous = "_synchronous=NORMAL"
	SYNC_FULL   Synchronous = "_synchronous=FULL"
	SYNC_EXTRA  Synchronous = "_synchronous=EXTRA"
)

type Mutex = Option

const (
	MUTEX_NO   Mutex = "_mutex=no"
	MUTEX_FULL Mutex = "_mutex=full"
)

type JournalMode = Option

const (
	JOURNAL_DELETE   JournalMode = "_journal_mode=DELETE"
	JOURNAL_TRUNCATE JournalMode = "_journal_mode=TRUNCATE"
	JOURNAL_PERSIST  JournalMode = "_journal_mode=PERSIST"
	JOURNAL_MEMORY   JournalMode = "_journal_mode=MEMORY"
	JOURNAL_WAL      JournalMode = "_journal_mode=WAL"
	JOURNAL_OFF      JournalMode = "_journal_mode=OFF"
)

type LockingMode = Option

const (
	LOCKING_NORMAL    LockingMode = "_locking_mode=NORMAL"
	LOCKING_EXCLUSIVE LockingMode = "_locking_mode=EXCLUSIVE"
)

type Mode = Option

const (
	MODE_READONLY        Mode = "mode=ro"
	MODE_READWRITE       Mode = "mode=rw"
	MODE_READWRITECREATE Mode = "mode=rwc"
	MODE_MEMORY          Mode = "mode=memory"
)

type ForeignKeys = Option

const (
	FK_ENABLED  ForeignKeys = "_fk=true"
	FK_DISABLED ForeignKeys = "_fk=false"
)
