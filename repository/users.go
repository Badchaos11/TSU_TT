package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Badchaos11/TSU_TT/model"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func (r *Repo) CreateUser(ctx context.Context, u model.User) (int64, error) {
	const query = `INSERT INTO users (name, surname, patronymic, sex, status, birth_date, created) VALUES ($1, $2, $3, $4, $5, $6, now()) RETURNING id;`
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	var id int64

	err := r.PGXRepo.QueryRow(ctx, query, u.Name, u.Surname, u.Patronymic, u.Sex, u.Status, u.BirthDate).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repo) ChangeUser(ctx context.Context, u model.ChangeUserRequest) (bool, error) {
	ub := squirrel.Update("users").PlaceholderFormat(squirrel.Dollar).Where(squirrel.Eq{"id": u.Id})
	fvMap := makeFieldValMap(u)
	for k, v := range fvMap {
		if v != "" {
			ub = ub.Set(k, v)
		}
	}
	exists, err := r.CheckIsUserExists(ctx, u.Id)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	sql, args, _ := ub.ToSql()
	_, err = r.PGXRepo.Exec(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Error("There is no user with this user ID %d", u.Id)
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *Repo) DeleteUser(ctx context.Context, userId int64) (bool, error) {
	const query = `DELETE FROM users WHERE id=$1`
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	exists, err := r.CheckIsUserExists(ctx, userId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	_, err = r.PGXRepo.Exec(ctx, query, userId)
	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Errorf("There is no user with id %d", userId)
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *Repo) GetUserByID(ctx context.Context, userId int) (*model.User, error) {
	const query = `SELECT name, surname, patronymic, sex, status, birth_date, created FROM users WHERE id=$1`
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	var u model.User

	uString, err := r.KVRepo.GetUsersFromCache(ctx, fmt.Sprintf("user-id:%d", userId))
	if err != nil {
		logrus.Errorf("Error getting user from cache: %v", err)
	}

	if uString != "" {
		err := json.Unmarshal([]byte(uString), &u)
		if err != nil {
			logrus.Errorf("Error unmarshalling user: %v", err)
			return nil, err
		}
		logrus.Info("Got user from cache")
		return &u, nil
	}

	logrus.Info("Couldn't find user in cache, go to db")

	err = r.PGXRepo.QueryRow(ctx, query, userId).Scan(&u.Name, &u.Surname, &u.Patronymic, &u.Sex, &u.Status, &u.BirthDate, &u.Created)
	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Errorf("There is no user with user ID %d", userId)
			return nil, nil
		}
		return nil, err
	}

	uStr, err := json.Marshal(u)
	if err != nil {
		logrus.Errorf("Error marshalling user %v", err)
		return &u, nil
	}

	err = r.KVRepo.AddToCache(ctx, fmt.Sprintf("user-id:%d", userId), string(uStr))
	if err != nil {
		logrus.Errorf("Error adding user to cache: %v", err)
	}

	return &u, nil
}

func (r *Repo) GetUsersFiltered(ctx context.Context, filter model.UserFilter) ([]model.User, error) {
	sq := squirrel.Select("name", "surname", "patronymic", "sex", "status", "birth_date", "created").From("users").PlaceholderFormat(squirrel.Dollar)
	if filter.Limit > 0 {
		sq = sq.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		sq = sq.Offset(filter.Offset)
	}

	if filter.Sex != "" {
		sq = sq.Where(squirrel.Eq{"sex": filter.Sex})
	}
	if filter.Status != "" {
		sq = sq.Where(squirrel.Eq{"status": filter.Status})
	}

	if filter.OrderBy == "sex" || filter.OrderBy == "status" {
		if filter.Desc != nil && *filter.Desc {
			sq = sq.OrderBy(filter.OrderBy)
			sq = sq.Suffix("desc")
		}
		sq = sq.OrderBy(filter.OrderBy)
	}

	if filter.Name != "" {
		sq = sq.Where(squirrel.Eq{"name": filter.Name})
	}
	if filter.Surname != "" {
		sq = sq.Where(squirrel.Eq{"surname": filter.Surname})
	}
	if filter.Patronymic != "" {
		sq = sq.Where(squirrel.Eq{"patronymic": filter.Patronymic})
	}

	sql, args, _ := sq.ToSql()
	var res []model.User

	uString, err := r.KVRepo.GetUsersFromCache(ctx, fmt.Sprintf("filtered-users:%s:%s:%s:%s:%s:%s:%v:%d:%d", filter.Sex, filter.Status, filter.Name, filter.Surname, filter.Patronymic, filter.OrderBy, *filter.Desc, filter.Limit, filter.Offset))
	if err != nil {
		logrus.Errorf("Error getting users from cache: %v", err)
	}

	if uString != "" {
		err := json.Unmarshal([]byte(uString), &res)
		if err != nil {
			logrus.Errorf("Error unmarshalling users: %v", err)
			return nil, err
		}
		logrus.Info("Got user from cache")
		return res, nil
	}

	logrus.Info("Can't find users in cache, go to db")
	rows, err := r.PGXRepo.Query(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Error("No userf for this filter")
			return nil, nil
		}
		logrus.Errorf("error select users error %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row model.User
		err := rows.Scan(&row.Name, &row.Surname, &row.Patronymic, &row.Sex, &row.Status, &row.BirthDate, &row.Created)
		if err != nil {
			logrus.Errorf("error scannig row %v", err)
			return nil, err
		}
		res = append(res, row)
	}

	uStr, err := json.Marshal(res)
	if err != nil {
		logrus.Errorf("Error marshalling user %v", err)
		return res, nil
	}

	err = r.KVRepo.AddToCache(ctx, fmt.Sprintf("filtered-users:%s:%s:%s:%s:%s:%s:%v:%d:%d", filter.Sex, filter.Status, filter.Name, filter.Surname, filter.Patronymic, filter.OrderBy, *filter.Desc, filter.Limit, filter.Offset), string(uStr))
	if err != nil {
		logrus.Errorf("Error adding user to cache: %v", err)
	}

	return res, nil
}

func (r *Repo) CheckIsUserExists(ctx context.Context, userId int64) (bool, error) {
	const query = `SELECT id FROM users WHERE id=$1`
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	var id int64

	err := r.PGXRepo.QueryRow(ctx, query, userId).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
