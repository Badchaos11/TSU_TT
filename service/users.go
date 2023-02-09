package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Badchaos11/TSU_TT/model"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func (s *service) CreateNewUser(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Got CreateNewUser Request. Starting process.")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("Error reading request body error %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось создать подьзователя из-за внутренней ошибки")
		return
	}
	var req model.User
	err = jsoniter.Unmarshal(body, &req)
	if err != nil {
		logrus.Errorf("Error unmarshalling request body error %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось создать пользователя из-за внутренней ошибки")
		return
	}
	if req.Name == "" || req.Surname == "" || req.Sex == "" {
		logrus.Errorf("empty one  or more requiered field")
		s.WriteResponse(w, http.StatusBadRequest, "Поля name, surname, sex должны быть обязательно заполнены")
	}

	req.Status = "Активен"
	ctx := context.Background()
	id, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		logrus.Errorf("Error creqting user error %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	logrus.Info("User succesfully created")
	s.WriteResponse(w, http.StatusOK, fmt.Sprintf("Пользователь успешно создан. ID пользователя %d", id))
}

func (s *service) CreateUsersFromExcell(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		logrus.Errorf("Error parsing multipart form: %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось прочитать файл")
		return
	}
	file, header, err := r.FormFile("user")
	if err != nil {
		logrus.Errorf("error reading file from multipart data %v :", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось прочитать данные из файла")
		return
	}

	user, err := s.GetUserFromFile(file, header.Size)
	if err != nil {
		logrus.Errorf("Error getting user from file %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось получить пользователя из файла")
		return
	}

	user.Status = "Активен"
	ctx := context.Background()
	id, err := s.repo.CreateUser(ctx, *user)
	if err != nil {
		logrus.Errorf("Error creating user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	s.WriteResponse(w, http.StatusOK, fmt.Sprintf("Пользователь успешно создан, id %d", id))
}

func (s *service) DeleteUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("error reading request body %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось удалить пользователя")
		return
	}
	var req model.DeleteUserRequest
	if err := jsoniter.Unmarshal(body, &req); err != nil {
		logrus.Errorf("error unmarshalling user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось удалить пользователя")
		return
	}
	if req.UserID <= 0 {
		logrus.Errorf("incorrect user id %v", req.UserID)
		s.WriteResponse(w, http.StatusBadRequest, "ID пользователя может быть только неотрицательным числом")
		return
	}
	ctx := context.Background()
	succes, err := s.repo.DeleteUser(ctx, req.UserID)
	if err != nil {
		logrus.Errorf("error deleting user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось удалить пользователя")
		return
	}
	if !succes {
		logrus.Errorf("can't find user with id %v", req.UserID)
		s.WriteResponse(w, http.StatusNotFound, "Пользователя с таким id не существует")
		return
	}

	s.WriteResponse(w, http.StatusOK, "Пользователь успешно удален")
}

func (s *service) ChangeUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("error reading request body %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось изменить данные пользователя")
		return
	}
	var req model.ChangeUserRequest

	if err := jsoniter.Unmarshal(body, &req); err != nil {
		logrus.Errorf("error unmarshalling request body %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось изменить данные пользователя")
		return
	}
	if req.Id <= 0 {
		logrus.Error("incorrect user id")
		s.WriteResponse(w, http.StatusBadRequest, "ID пользователя должен быть неотрицательным и быть больше 0")
	}

	ctx := context.Background()
	exists, err := s.repo.ChangeUser(ctx, req)
	if err != nil {
		logrus.Errorf("error updating user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось изменить данные пользователя")
		return
	}
	if !exists {
		logrus.Errorf("can't find user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Пользователя с таким id не существует")
		return
	}

	s.WriteResponse(w, http.StatusOK, "Данные пользователя успешно изменены")
}

func (s *service) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("user_id")
	if userIdStr == "" {
		logrus.Errorf("quiery user_id is empty")
		s.WriteResponse(w, http.StatusBadRequest, "Введен пустой id пользователя")
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		logrus.Errorf("entered incorrect user_id")
		s.WriteResponse(w, http.StatusBadRequest, "Введен некорректный id пользователя, можно использовать только цифры")
		return
	}
	if userId <= 0 {
		logrus.Errorf("entered incorrect user_id")
		s.WriteResponse(w, http.StatusBadRequest, "Введен некорректный id пользователя, можно использовать только числа больше 0")
		return
	}
	ctx := context.Background()
	user, err := s.repo.GetUserByID(ctx, userId)
	if err != nil {
		logrus.Errorf("error quering user %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось найти пользователя")
		return
	}
	if user == nil {
		logrus.Errorf("can't find user with user_id %d", userId)
		s.WriteResponse(w, http.StatusNotFound, "Пользователь с таким id не существует")
		return
	}

	strBody, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("error marshalling response %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось найти пользователя")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strBody))
}

func (s *service) GetFilteredUsers(w http.ResponseWriter, r *http.Request) {
	filter := s.GetUserFilter(r)

	ctx := context.Background()
	users, err := s.repo.GetUsersFiltered(ctx, filter)
	if err != nil {
		logrus.Errorf("error quering user filter: %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось найти пользователей по данному фильтру")
		return
	}

	if users == nil {
		logrus.Errorf("can't find users for this filter")
		s.WriteResponse(w, http.StatusNotFound, "Для данного фильтра пользователи не найдены")
		return
	}

	strBody, err := json.Marshal(users)
	if err != nil {
		logrus.Errorf("error marshalling users: %v", err)
		s.WriteResponse(w, http.StatusInternalServerError, "Не удалось найти пользователей по данному фильтру")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strBody))
}
