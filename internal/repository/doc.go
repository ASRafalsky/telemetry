/*
Package repository contains repository with database and cache extension,
and all necessary methods for interaction with it.

# TYPES

	type ExtendedRepository struct {
		// Has unexported fields.
	}

ExtendedRepository describes ExtendedRepository.

func NewExtendedRepository(cache cache) *ExtendedRepository
NewExtendedRepository creates new ExtendedRepository instance with cache.

func (r *ExtendedRepository) CacheSize() int
CacheSize returns the number of entries in the cache.

func (r *ExtendedRepository) Delete(ctx context.Context, key string) error
Delete removes entry from repository by key and returns error if anything
went wrong.

func (r *ExtendedRepository) ForEach(ctx context.Context, fn func(k string, v []byte) error) error
ForEach calls fn for each entry in the repository. And you know what? Yeah!
It returns error if anything went wrong.

func (r *ExtendedRepository) Get(ctx context.Context, key string) ([]byte, error)
Get returns data from repository and error if anything went wrong. If this
entry doesn't exist it returns nil, nil (I know, this is bad behavior,
don't follow me).

TODO(ASRafalsky): Do something about return nil, nil.

func (r *ExtendedRepository) IsReady() bool
IsReady returns true if ExtendedRepository contains new data.

func (r *ExtendedRepository) Maintain(ctx context.Context, cfg config.Server, l log.Logger)
Maintain maintains all repository processing. It syncs repository with DB
and do backup according to cfg.

func (r *ExtendedRepository) Ping(ctx context.Context) error
Ping checks db if it sets.

func (r *ExtendedRepository) Ready()
Ready set ExtendedRepository to the no new data state.

func (r *ExtendedRepository) Set(key string, value []byte)
Set sets key value pair.

func (r *ExtendedRepository) Size() (int, error)
Size returns the number of entries in the repository. If it uses db it will
be number entries from the db, else from the cache.

func (r *ExtendedRepository) Sync(ctx context.Context) error
Sync syncs cache with db.

func (r *ExtendedRepository) UseDB(db db)
UseDB adds db to the ExtendedRepository instance.
*/
package repository
