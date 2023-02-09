package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Badchaos11/TSU_TT/model"
	"github.com/Badchaos11/TSU_TT/repository"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	port string
	repo repository.IRepository
}

func NewService(ctx context.Context, config *model.Config) (*service, error) {
	dsnPg := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", config.DBUser, config.DBPassword, config.DBHost, config.DBName)
	conn, err := repository.NewRepository(ctx, dsnPg, os.Getenv("CACHE_URL"))
	if err != nil {
		logrus.Errorf("Unable to connect database error %v", err)
		return nil, err
	}
	return &service{port: config.Port, repo: conn}, nil
}

func (s *service) Run() {

	router := mux.NewRouter()

	router.HandleFunc("/get_user_by_id", s.GetUserByID).Methods("GET")
	router.HandleFunc("/get_filtered_users", s.GetFilteredUsers).Methods("GET")
	router.HandleFunc("/create_user", s.CreateNewUser).Methods("POST")
	router.HandleFunc("/create_user_from_file", s.CreateUsersFromExcell).Methods("POST")
	router.HandleFunc("/change_user", s.ChangeUser).Methods("POST")
	router.HandleFunc("/delete_user", s.DeleteUser).Methods("DELETE")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.port), // Порт сервера
		Handler:      router,                     // Хэндлеры
		ReadTimeout:  5 * time.Second,            // Таймаут запроса клиента
		WriteTimeout: 10 * time.Second,           // Таймаут ответа клиенту
		IdleTimeout:  120 * time.Second,          // Таймаут соединения в простое
	}

	go func() {
		logrus.Infof("starting server on port %v", s.port)

		err := server.ListenAndServe()
		if err != nil {
			logrus.Errorf("error starting server %v", err)
			os.Exit(1)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Minute * 30)
			logrus.Info("Cleaning cache...")
			ctx := context.Background()
			err := s.repo.ClearCache(ctx)
			if err != nil {
				logrus.Errorf("Error cleaning cache %v", err)
			}
		}
	}()

	// Отключение
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	sig := <-c
	logrus.Infof("Got signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.Shutdown(ctx)
}

func init() {

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
		FullTimestamp:   true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}
