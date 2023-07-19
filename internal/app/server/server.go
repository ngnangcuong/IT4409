package server

import (
	"IT4409/internal/app/database"
	"IT4409/internal/app/handler"
	blogRepository "IT4409/internal/app/repositories/blog"
	commentRepository "IT4409/internal/app/repositories/comment"
	permissionRepository "IT4409/internal/app/repositories/permission"
	tokenRepository "IT4409/internal/app/repositories/token"
	userRepository "IT4409/internal/app/repositories/user"
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
	redisClient := database.GetRedisClient(viper.GetString("database.redis.dsn"))
	blogRepo := blogRepository.NewBlogRepo(postgres)
	commentRepo := commentRepository.NewCommentRepo(postgres)
	userRepo := userRepository.NewUserRepo(postgres)
	tokenRepo := tokenRepository.NewTokenRepo(redisClient)
	permissionRepo := permissionRepository.NewPermissionRepo(postgres)
	blogService := services.NewBlogService(blogRepo, commentRepo, permissionRepo, userRepo, postgres)
	commentService := services.NewCommentService(commentRepo, permissionRepo, userRepo, postgres)
	userService := services.NewUserService(userRepo, postgres)
	tokenService := services.NewTokenService(tokenRepo, viper.GetInt64("app.at_expires"), viper.GetInt64("app.rt_expires"),
		viper.GetString("app.access_secret"), viper.GetString("app.refresh_secret"))
	blogHandler := handler.NewBlogHandler(blogService)
	commentHandler := handler.NewCommentHandler(commentService)
	authHandler := handler.NewAuthHandler(userService, tokenService)
	router := handler.GetRouter(blogHandler, commentHandler, authHandler)

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
