package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"github.com/satria/obrolan-api/internal/services"
	"github.com/satria/obrolan-api/internal/utils"
)

type ChatHandler struct {
	hub      *services.Hub
	msgRepo  *repositories.MessageRepository
	cfg      *config.Config
	upgrader websocket.Upgrader
}

func NewChatHandler(hub *services.Hub, msgRepo *repositories.MessageRepository, cfg *config.Config) *ChatHandler {
	allowedOrigins := cfg.CORSOrigins

	return &ChatHandler{
		hub:     hub,
		msgRepo: msgRepo,
		cfg:     cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, a := range allowedOrigins {
					if origin == a {
						return true
					}
				}
				return false
			},
		},
	}
}

// HandleWebSocket godoc
// @Summary      WebSocket chat
// @Description  Upgrade to WebSocket for real-time chat in a thread
// @Tags         Chat
// @Produce      json
// @Param        id path string true "Thread ID (UUID)"
// @Param        token query string true "JWT token"
// @Success      200 {object} utils.APIResponse
// @Router       /api/v1/threads/{id}/chat [get]
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.ErrorJSON(c, http.StatusBadRequest, "thread id is required")
		return
	}

	if _, err := uuid.Parse(threadID); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	token := c.Query("token")
	if token == "" {
		utils.ErrorJSON(c, http.StatusUnauthorized, "missing token")
		return
	}

	claims, err := utils.ValidateToken(h.cfg, token)
	if err != nil {
		utils.ErrorJSON(c, http.StatusUnauthorized, "invalid token")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := services.NewClient(h.hub, conn, claims.UserID, claims.Username, threadID, func(msg *models.Message) error {
		return h.msgRepo.Create(msg)
	})

	h.hub.RegisterCh <- client

	go client.WritePump()
	go client.ReadPump()
}

type MessageResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
}

// GetMessages godoc
// @Summary      Get chat messages
// @Description  Get message history for a thread
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Param        limit query int false "Number of messages" default(50)
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/threads/{id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	messages, err := h.msgRepo.GetByThread(threadID, limit)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		username := ""
		avatarURL := ""
		if msg.User.ID != uuid.Nil {
			username = msg.User.Username
			avatarURL = msg.User.AvatarURL
		}

		responses[i] = MessageResponse{
			ID:        msg.ID,
			UserID:    msg.UserID,
			Username:  username,
			AvatarURL: avatarURL,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format(utils.TimeFormat),
		}
	}

	utils.SuccessJSON(c, responses)
}
