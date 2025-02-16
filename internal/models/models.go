package models

import "time"

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Coins    int    `db:"coins"`
}

type CoinTransaction struct {
	ID         int       `db:"id"`
	FromUserID *int      `db:"from_user_id"` 
	ToUserID   *int      `db:"to_user_id"`   
	Amount     int       `db:"amount"`
	CreatedAt  time.Time `db:"created_at"`
}

type ItemPurchase struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	ItemName  string    `db:"item_name"`
	Quantity  int       `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
}


type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type InfoResponse struct {
	Coins       int           `json:"coins"`
	Inventory   []InvItem     `json:"inventory"`
	CoinHistory CoinHistory   `json:"coinHistory"`
}

type InvItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedCoin `json:"received"`
	Sent     []SentCoin     `json:"sent"`
}

type ReceivedCoin struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentCoin struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}
