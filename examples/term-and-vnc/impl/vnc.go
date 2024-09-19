package impl

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luthermonson/go-proxmox"
	"github.com/rs/zerolog/log"
)

var tickets = make(map[string]*proxmox.VNC)

func VncTicket(c *gin.Context) {
	vm, err := GetVm()
	if err != nil {
		log.Error().Err(err).Msg("Error getting version")
		return
	}

	vnc, err := vm.VNCProxy(context.Background(), nil)
	if err != nil {
		log.Error().Err(err).Msg("Error getting version")
		return
	}

	// log.Debug().Str("Ticket", vnc.Ticket).Send()
	id := base64.StdEncoding.EncodeToString([]byte(vnc.Ticket))
	tickets[id] = vnc

	c.JSON(http.StatusOK, gin.H{"ticket": id})
}

func Vnc(c *gin.Context) {
	log.Debug().Msg("New VNC connection")
	// vnc, err := vm.VNCProxy(context.Background(), nil)
	id := c.Query("ticket")
	vm, err := GetVm()
	if err != nil {
		log.Error().Err(err).Msg("Error getting version")
		return
	}

	// vnc, err := vm.VNCProxy(context.Background(), nil)
	// if err != nil {
	// 	log.Error().Err(err).Msg("Error getting version")
	// 	return
	// }
	vnc := tickets[string(id)]

	if vnc == nil {
		log.Error().Msg("VNC ticket not found")
		return
	}

	// log.Debug().Str("Ticket", vnc.Ticket).Send()
	send, recv, errs, close, err := vm.VNCWebSocket(vnc)
	if err != nil {
		log.Error().Err(err).Msg("Error getting version")
		return
	}
	defer close()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-recv:
				// log.Debug().Bytes("msg", msg).Msg("proxmox:")
				err = ws.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					done <- struct{}{}
					log.Error().Err(err).Send()
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case err := <-errs:
				if err != nil {
					log.Error().Err(err).Send()
				}
				done <- struct{}{}
				return
			default:
				_, msg, err := ws.ReadMessage()
				if err != nil {
					done <- struct{}{}
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						return
					}
					log.Error().Err(err).Msg("Error reading from websocket")
					return
				}
				// log.Debug().Bytes("msg", msg).Msg("Client:")
				send <- msg
			}
		}
	}()

	<-done
}
