package controllers

import (
	"log"
	"net/http"
	"shwetaik-expense-management-api/configs"
	"strconv"

	"github.com/labstack/echo/v4"
)

func HandleWebSocket(c echo.Context) error {
	userId := c.Param("userID")
	if userId == "" {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user id")
	}

	ws, err := configs.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer ws.Close()

	configs.WebSocketConnections.Lock()
	configs.WebSocketConnections.M[uint(id)] = &configs.WebSocketConnection{Conn: ws}
	configs.WebSocketConnections.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			configs.WebSocketConnections.Lock()
			delete(configs.WebSocketConnections.M, uint(id))
			configs.WebSocketConnections.Unlock()
			break
		}
	}

	return nil
}
