package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Badchaos11/TSU_TT/model"
	"github.com/Badchaos11/TSU_TT/repository/cache"
	"github.com/fatih/structs"
	"github.com/jackc/pgx/v5"
)

type IRepository interface {
	CreateUser(ctx context.Context, u model.User) (int64, error)
	ChangeUser(ctx context.Context, u model.ChangeUserRequest) (bool, error)
	DeleteUser(ctx context.Context, userId int64) (bool, error)
	GetUserByID(ctx context.Context, userId int) (*model.User, error)
	GetUsersFiltered(ctx context.Context, filter model.UserFilter) ([]model.User, error)
	CheckIsUserExists(ctx context.Context, userId int64) (bool, error)
	ClearCache(ctx context.Context) error
}

const initSqlTable = `CREATE TABLE IF NOT EXISTS users (
	id bigint primary key generated by default as identity,
	name text not null,
	surname text not null,
	patronymic text,
	sex text not null,
	status text not null,
	birth_date timestamp,
	created timestamp not null
);`

type Repo struct {
	PGXRepo *pgx.Conn
	KVRepo  cache.ICacheRepository
	timeout time.Duration
}

func NewRepository(ctx context.Context, dsn string, cacheUrl string) (IRepository, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	conn.Exec(ctx, initSqlTable)

	cacheConn, err := cache.NewCacheClient(ctx, cacheUrl)
	if err != nil {
		return nil, err
	}

	return &Repo{
		PGXRepo: conn,
		KVRepo:  cacheConn,
		timeout: time.Second * 5,
	}, nil
}

func makeFieldValMap(u model.ChangeUserRequest) map[string]string {
	fields := structs.Fields(u)
	res := make(map[string]string, 0)

	for _, field := range fields {
		f := field.Tag("json")
		f = strings.Split(f, ",")[0]
		if f == "id" || f == "birth_date" {
			continue
		}
		res[f] = field.Value().(string)
	}

	return res
}

func (r *Repo) ClearCache(ctx context.Context) error {
	return r.KVRepo.ClearCache(ctx)
}
