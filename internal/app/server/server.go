package server

import (
	"IT4409/internal/app/database"
	"IT4409/internal/app/handler"
	blogRepository "IT4409/internal/app/repositories/blog"
	commentRepository "IT4409/internal/app/repositories/comment"
	"IT4409/internal/app/services"
	"fmt"

	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

func Run() {
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}
	database.InitPostgresql(viper.GetString("database.postgresql.host"), viper.GetString("database.postgresql.user"),
		viper.GetString("database.postgresql.password"), viper.GetString("database.postgresql.dbname"), viper.GetInt("database.postgresql.port"))
	postgres := database.GetConnectionPool()
	defer postgres.Close()
	blogRepo := blogRepository.NewBlogRepo(postgres)
	commentRepo := commentRepository.NewCommentRepo(postgres)
	blogService := services.NewBlogService(blogRepo, commentRepo, postgres)
	commentService := services.NewCommentService(commentRepo, postgres)
	blogHandler := handler.NewBlogHandler(blogService)
	commentHandler := handler.NewCommentHandler(commentService)
	router := handler.GetRouter(blogHandler, commentHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	serveHTTP := func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := srv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go serveHTTP(&wg)
	wg.Wait()
}
