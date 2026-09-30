//go:generate oapi-codegen -config=../openapi/config.yaml ../openapi/openapi.yaml
package application

import (
	"net/http"

	"github.com/Pluslab/cyphonic/noded/application/daemon"
	"github.com/Pluslab/cyphonic/noded/entity/service"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/gin-gonic/gin"
)

var _ ServerInterface = (*server)(nil)

type server struct {
	authService service.AuthService
	daemon      daemon.Daemon
}

func NewServer(authService service.AuthService, daemon daemon.Daemon) *server {
	return &server{
		authService: authService,
		daemon:      daemon,
	}
}

func (s *server) PostLogin(ginContext *gin.Context) {
	config.LogInfo("Begin login process")
	defer config.LogInfo("End login process")

	var body PostLoginJSONBody
	if err := ginContext.ShouldBindJSON(&body); err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if body.Username == "" || body.Password == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})

		return
	}

	// TODO: Implement Digital Signature
	// TODO: Implement Single Sign-On
	message, loginResponse, err := s.authService.Login(body.Username, body.Password)
	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := s.daemon.Provision(*loginResponse); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := s.daemon.Connection(*loginResponse); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := s.daemon.Registration(*loginResponse); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := s.daemon.Up(); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ginContext.JSON(http.StatusOK, gin.H{"message": message})
}

func (s *server) PostLogout(ginContext *gin.Context) {
	config.LogInfo("Begin logout process")
	defer config.LogInfo("End logout process")

	if err := s.daemon.Finalization(s.daemon.LoginResponse()); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	if err := s.daemon.Down(); err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ginContext.JSON(http.StatusOK, gin.H{"message": "logout successful!"})
}

func (s *server) GetFqdn(ginContext *gin.Context) {
	config.LogInfo("Begin get FQDN process")
	defer config.LogInfo("End get FQDN process")

	fqdn := s.daemon.GetFqdn()
	if fqdn == "" {
		ginContext.JSON(http.StatusNotFound, gin.H{"error": "fqdn not found"})
	}

	ginContext.JSON(http.StatusOK, gin.H{"fqdn": fqdn})
}

func (s *server) GetStatus(ginContext *gin.Context) {
	config.LogInfo("Begin get status process")
	defer config.LogInfo("End get status process")

	statusMessage := s.daemon.GetStatus()
	ginContext.JSON(http.StatusOK, gin.H{"status": statusMessage})
}
