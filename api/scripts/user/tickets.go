package user

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type TicketRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type Ticket struct {
	ID        string `json:"ticket_id"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateTicketHandler creates a new support ticket.
func CreateTicketHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req TicketRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Subject == "" || req.Message == "" {
		utils.LogActivity(c, uid, "create_ticket", false, "invalid payload")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "données invalides"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "create_ticket", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	id, _ := uuid.NewV7()
	ticketID := id.String()

	_, err = db.Exec(`INSERT INTO support_tickets (ticket_id, user_id, subject, message) VALUES (?,?,?,?)`,
		ticketID, uid, req.Subject, req.Message)
	if err != nil {
		utils.LogActivity(c, uid, "create_ticket", false, "insert error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, uid, "create_ticket", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "ticket_id": ticketID})
}

// ListTicketsHandler returns tickets for the authenticated user.
func ListTicketsHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "list_tickets", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT ticket_id, subject, message, status, created_at, updated_at FROM support_tickets WHERE user_id = ? ORDER BY created_at DESC`, uid)
	if err != nil {
		utils.LogActivity(c, uid, "list_tickets", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	tickets := make([]Ticket, 0)
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.Subject, &t.Message, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			utils.LogActivity(c, uid, "list_tickets", false, "scan error")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		tickets = append(tickets, t)
	}

	utils.LogActivity(c, uid, "list_tickets", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "tickets": tickets})
}
