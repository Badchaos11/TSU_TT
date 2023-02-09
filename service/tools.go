package service

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/Badchaos11/TSU_TT/model"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
)

func (s *service) WriteResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	type response struct {
		Message string `json:"message"`
	}
	res := response{Message: msg}
	body, _ := json.Marshal(&res)
	w.Write(body)
}

func (s *service) GetUserFilter(r *http.Request) model.UserFilter {
	var limit, offset int64
	var desc bool

	sex := r.URL.Query().Get("sex")
	status := r.URL.Query().Get("status")
	name := r.URL.Query().Get("name")
	surname := r.URL.Query().Get("surname")
	patr := r.URL.Query().Get("patronymic")
	orderBy := r.URL.Query().Get("order_by")

	descStr := r.URL.Query().Get("desc")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = 0
	}
	offset, err = strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		offset = 0
	}

	if descStr != "" {
		desc, _ = strconv.ParseBool(descStr)
	}

	return model.UserFilter{
		Sex:        sex,
		Status:     status,
		OrderBy:    orderBy,
		Limit:      uint64(limit),
		Offset:     uint64(offset),
		Desc:       &desc,
		Name:       name,
		Surname:    surname,
		Patronymic: patr,
	}
}

func (s *service) GetUserFromFile(file multipart.File, size int64) (*model.User, error) {
	f, err := xlsx.OpenReaderAt(file, size)
	if err != nil {
		return nil, err
	}

	ss := f.Sheets[0]
	name := ss.Cell(0, 0).Value
	surname := ss.Cell(1, 0).Value
	patr := ss.Cell(2, 0).Value
	sex := ss.Cell(3, 0).Value
	email := ss.Cell(4, 0).Value

	logrus.Infof("parsed email from file %s", email)

	return &model.User{
		Name:       name,
		Surname:    surname,
		Patronymic: patr,
		Sex:        sex,
	}, nil
}
