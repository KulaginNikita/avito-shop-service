package repository

import (
	"database/sql"
	"fmt"

	"avito-shop/internal/config"
	"avito-shop/internal/models"

	_ "github.com/lib/pq"
)

type Repository interface {
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(username, password string) (int, error)
	UpdateUserCoins(userID, newAmount int) error

	InsertCoinTransaction(fromUserID, toUserID *int, amount int) error
	GetCoinTransactionsByUserID(userID int) ([]models.CoinTransaction, error)

	InsertItemPurchase(userID int, itemName string, quantity int) error
	GetAllPurchasesByUserID(userID int) ([]models.ItemPurchase, error)

	GetUserByID(userID int) (*models.User, error)
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func NewRepository(db *sql.DB) Repository {
	return &PostgresRepo{
		db: db,
	}
}


func (r *PostgresRepo) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password, coins FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Coins)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepo) CreateUser(username, password string) (int, error) {
	query := `INSERT INTO users (username, password, coins) VALUES ($1, $2, 1000) RETURNING id`
	var id int
	err := r.db.QueryRow(query, username, password).Scan(&id)
	return id, err
}

func (r *PostgresRepo) UpdateUserCoins(userID, newAmount int) error {
	query := `UPDATE users SET coins = $1 WHERE id = $2`
	_, err := r.db.Exec(query, newAmount, userID)
	return err
}

func (r *PostgresRepo) InsertCoinTransaction(fromUserID, toUserID *int, amount int) error {
	query := `INSERT INTO coin_transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, fromUserID, toUserID, amount)
	return err
}

func (r *PostgresRepo) GetCoinTransactionsByUserID(userID int) ([]models.CoinTransaction, error) {
	query := `SELECT id, from_user_id, to_user_id, amount, created_at 
			  FROM coin_transactions
			  WHERE from_user_id = $1 OR to_user_id = $1
			  ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.CoinTransaction
	for rows.Next() {
		var c models.CoinTransaction
		if err := rows.Scan(&c.ID, &c.FromUserID, &c.ToUserID, &c.Amount, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *PostgresRepo) InsertItemPurchase(userID int, itemName string, quantity int) error {
	query := `INSERT INTO item_purchases (user_id, item_name, quantity) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, userID, itemName, quantity)
	return err
}

func (r *PostgresRepo) GetAllPurchasesByUserID(userID int) ([]models.ItemPurchase, error) {
	query := `SELECT id, user_id, item_name, quantity, created_at 
			  FROM item_purchases WHERE user_id = $1 
			  ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []models.ItemPurchase
	for rows.Next() {
		var ip models.ItemPurchase
		if err := rows.Scan(&ip.ID, &ip.UserID, &ip.ItemName, &ip.Quantity, &ip.CreatedAt); err != nil {
			return nil, err
		}
		purchases = append(purchases, ip)
	}
	return purchases, nil
}

func (r *PostgresRepo) GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, username, password, coins FROM users WHERE id = $1`
	row := r.db.QueryRow(query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Coins)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
