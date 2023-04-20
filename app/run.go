package main

import (
	"core/core/pkg/grpc/implementation"
	"core/core/pkg/grpc/protobuf"
	"core/core/pkg/models"
	"core/core/pkg/rest"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Steps:
	// 1. Pull the configuration values from the configuration file
	// 2. Connect to the database
	// 3. Connect to the task queue
	// 4. Run Tests?
	// 5. Run Gin Webserver
	// 6. Start GRPC Server

	// Fetching Config
	fmt.Println("Fetching Config")
	config, err := ConfigurationFromConfig()
	if err != nil {
		fmt.Print(err)
	}

	// Connecting to the configured database

	fmt.Println("Connecting Database")
	connectionString := "postgres://" + config.Database.User + ":" + config.Database.Pass + "@" + config.Database.Host + "/" + config.Database.Name
	database, err := gorm.Open(
		postgres.Open(connectionString),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		fmt.Print(err)
	}
	database.AutoMigrate(
		models.Account{},
		models.PaymentRequest{},
	)
	// Connecting to the configured Redis task queue

	fmt.Println("Connecting Task Queue")
	redisQueue := redis.NewClient(
		&redis.Options{
			Addr:     config.TaskQueue.Host,
			Username: config.TaskQueue.User,
			Password: config.TaskQueue.Pass,
			DB:       config.TaskQueue.DB,
		},
	)
	var QueueFactory = redisq.NewFactory()
	var mainQueue = QueueFactory.RegisterQueue(
		&taskq.QueueOptions{
			Name:  "general-worker",
			Redis: redisQueue,
		},
	)
	mainQueue.Close()
	fmt.Print(database.DryRun)
	// Setting up Gin Webserver

	var g errgroup.Group

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return server.ListenAndServe()
	})

	// Setting up GRPC server
	go StartGrpcServer(*database, mainQueue)

	// Waiting on the Webserver
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

// Add newly implemented routes here
func router() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	rest.InitializePingRoutes(e)
	return e
}

func StartGrpcServer(database gorm.DB, taskqueue taskq.Queue) {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	protobuf.RegisterCoreServer(s, &implementation.Server{
		Database:  database,
		TaskQueue: taskqueue,
	})
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type Configuration struct {
	Database  DatabaseConfiguration
	TaskQueue TaskQueueConfiguration
}

type DatabaseConfiguration struct {
	User string
	Name string
	Pass string
	Host string
	Port int
}

type TaskQueueConfiguration struct {
	Host string
	User string
	Pass string
	DB   int
}

func ConfigurationFromConfig() (*Configuration, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
		return nil, err
	}
	err := viper.Unmarshal(&Configuration{})
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
	var databaseConfig DatabaseConfiguration = DatabaseConfiguration{
		User: viper.GetString("database.user"),
		Host: viper.GetString("database.host"),
		Pass: viper.GetString("database.pass"),
		Name: viper.GetString("database.name"),
		Port: viper.GetInt("database.port"),
	}
	var taskQueueConfig TaskQueueConfiguration = TaskQueueConfiguration{
		Host: viper.GetString("task_queue.host"),
		DB:   viper.GetInt("task_queue.db"),
		Pass: viper.GetString("task_queue.pass"),
		User: viper.GetString("task_queue.user"),
	}
	return &Configuration{
		Database:  databaseConfig,
		TaskQueue: taskQueueConfig,
	}, nil
}
