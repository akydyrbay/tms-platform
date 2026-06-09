package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"

	"tms-platform/internal/auth"
	"tms-platform/internal/database"
	"tms-platform/internal/folder"
	"tms-platform/internal/project"
	"tms-platform/internal/testcase"
	"tms-platform/internal/user"
)

type Server struct {
	port     int
	db       database.Service
	jwt      *auth.JWTManager
	auth     *auth.AuthHandler
	project  *project.Handler
	folder   *folder.Handler
	testcase *testcase.Handler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()

	dbx := sqlx.NewDb(db.DB(), "pgx")

	jwtMgr := auth.NewJWTManager([]byte(os.Getenv("JWT_SECRET")), 30*time.Minute)
	authSvc := auth.NewAuthService(user.NewUserRepository(dbx), jwtMgr)

	projectSvc := project.NewService(project.NewRepository(dbx))
	folderSvc := folder.NewService(folder.NewRepository(dbx))
	testcaseSvc := testcase.NewService(testcase.NewRepository(dbx))

	NewServer := &Server{
		port:     port,
		db:       db,
		jwt:      jwtMgr,
		auth:     auth.NewAuthHandler(authSvc),
		project:  project.NewHandler(projectSvc),
		folder:   folder.NewHandler(folderSvc),
		testcase: testcase.NewHandler(testcaseSvc),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
