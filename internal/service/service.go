package service

import (
	"errors"
	"strings"
	"time"
	"fmt"
	"avito-shop/internal/config"
	"avito-shop/internal/models"
	"avito-shop/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrNotEnoughCoins  = errors.New("not enough coins")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidItem     = errors.New("invalid item")
	ErrNegativeAmount  = errors.New("amount must be positive")
)

var itemPrices = map[string]int{
    "t-shirt":    80,
    "cup":        20,
    "book":       50,
    "pen":        10,
    "powerbank":  200,
    "hoody":      300,
    "umbrella":   200,
    "socks":      10,
    "wallet":     50,
    "pink-hoody": 500,
}

type Service interface {
    AuthUser(username, password string) (string, error)
    GetInfo(userID int) (*models.InfoResponse, error)
    SendCoin(fromUserID int, toUsername string, amount int) error
    BuyItem(userID int, itemName string) error
}

type service struct {
    repo repository.Repository
    cfg  *config.Config
}

func NewService(repo repository.Repository, cfg *config.Config) Service {
    return &service{repo: repo, cfg: cfg}
}

// ----------------------------------------
// AuthUser
// ----------------------------------------
func (s *service) AuthUser(username, password string) (string, error) {
    username = strings.TrimSpace(username)
    if username == "" || password == "" {
        return "", errors.New("invalid password")
    }

    user, err := s.repo.GetUserByUsername(username)
    if err != nil {
        return "", err
    }

    if user == nil {
        hashedPass, errHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if errHash != nil {
            return "", errHash
        }
        newUserID, errCreate := s.repo.CreateUser(username, string(hashedPass))
        if errCreate != nil {
            return "", errCreate
        }
        token, errToken := GenerateJWT(newUserID, s.cfg.JWTSecret)
        return token, errToken
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return "", errors.New("invalid password")
    }

    token, err := GenerateJWT(user.ID, s.cfg.JWTSecret)
    return token, err
}

// ----------------------------------------
// GetInfo
// ----------------------------------------

func (s *service) GetInfo(userID int) (*models.InfoResponse, error) {
    user, err := s.repo.GetUserByID(userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }

    purchases, err := s.repo.GetAllPurchasesByUserID(user.ID)
    if err != nil {
        return nil, err
    }
    inventoryMap := make(map[string]int)
    for _, p := range purchases {
        inventoryMap[p.ItemName] += p.Quantity
    }

    var inventory []models.InvItem
    for k, v := range inventoryMap {
        inventory = append(inventory, models.InvItem{Type: k, Quantity: v})
    }

    transactions, err := s.repo.GetCoinTransactionsByUserID(user.ID)
    if err != nil {
        return nil, err
    }

    received := make([]models.ReceivedCoin, 0)
    sent := make([]models.SentCoin, 0)

    for _, t := range transactions {
        if t.ToUserID != nil && *t.ToUserID == user.ID {
            fromName := "store"
            if t.FromUserID != nil {
                fromUser, _ := s.repo.GetUserByID(*t.FromUserID)
                if fromUser != nil {
                    fromName = fromUser.Username
                }
            }
            received = append(received, models.ReceivedCoin{
                FromUser: fromName,
                Amount:   t.Amount,
            })
        } else if t.FromUserID != nil && *t.FromUserID == user.ID {
            toName := "store"
            if t.ToUserID != nil {
                toUser, _ := s.repo.GetUserByID(*t.ToUserID)
                if toUser != nil {
                    toName = toUser.Username
                }
            }
            sent = append(sent, models.SentCoin{
                ToUser: toName,
                Amount: t.Amount,
            })
        }
    }

    return &models.InfoResponse{
        Coins:     user.Coins,
        Inventory: inventory,
        CoinHistory: models.CoinHistory{
            Received: received,
            Sent:     sent,
        },
    }, nil
}


// ----------------------------------------
// SendCoin
// ----------------------------------------

func (s *service) SendCoin(fromUserID int, toUsername string, amount int) error {
    // LOG: выводим параметры
    fmt.Printf("SendCoin: fromUserID=%d, toUser=%s, amount=%d\n", fromUserID, toUsername, amount)

    if amount <= 0 {
        return errors.New("amount must be positive")
    }
    toUsername = strings.TrimSpace(toUsername)
    if toUsername == "" {
        return errors.New("empty toUser")
    }

    fromUser, err := s.repo.GetUserByID(fromUserID)
    if err != nil {
        return err
    }
    if fromUser == nil {
        return errors.New("user not found")
    }

    fmt.Printf("SendCoin: fromUser before => ID=%d, coins=%d\n", fromUser.ID, fromUser.Coins)

    toUser, err := s.repo.GetUserByUsername(toUsername)
    if err != nil {
        return err
    }
    if toUser == nil {
        return errors.New("recipient not found")
    }

    fmt.Printf("SendCoin: toUser before => ID=%d, coins=%d\n", toUser.ID, toUser.Coins)

    if fromUser.Coins < amount {
        return errors.New("not enough coins")
    }

    fromUser.Coins -= amount
    toUser.Coins += amount

    fmt.Printf("SendCoin: fromUser after => ID=%d, coins=%d\n", fromUser.ID, fromUser.Coins)
    fmt.Printf("SendCoin: toUser after => ID=%d, coins=%d\n", toUser.ID, toUser.Coins)

    if err := s.repo.UpdateUserCoins(fromUser.ID, fromUser.Coins); err != nil {
        return err
    }
    if err := s.repo.UpdateUserCoins(toUser.ID, toUser.Coins); err != nil {
        return err
    }

    err = s.repo.InsertCoinTransaction(&fromUserID, &toUser.ID, amount)
    if err != nil {
        return err
    }

    return nil
}


// ----------------------------------------
// BuyItem
// ----------------------------------------

func (s *service) BuyItem(userID int, itemName string) error {
    itemName = strings.TrimSpace(itemName)
    price, ok := itemPrices[itemName]
    if !ok {
        return errors.New("invalid item")
    }

    fmt.Printf("BuyItem: userID=%d, itemName=%s, price=%d\n", userID, itemName, price)

    user, err := s.repo.GetUserByID(userID)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }

    fmt.Printf("BuyItem: user before => ID=%d, coins=%d\n", user.ID, user.Coins)

    if user.Coins < price {
        return errors.New("not enough coins")
    }

    newCoins := user.Coins - price

    fmt.Printf("BuyItem: user after => ID=%d, coins=%d\n", user.ID, newCoins)

    if err := s.repo.UpdateUserCoins(user.ID, newCoins); err != nil {
        return err
    }

    if err := s.repo.InsertItemPurchase(user.ID, itemName, 1); err != nil {
        return err
    }

    err = s.repo.InsertCoinTransaction(&userID, nil, price)
    if err != nil {
        return err
    }

    return nil
}


// ----------------------------------------
// GenerateJWT
// ----------------------------------------

func GenerateJWT(userID int, secret string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     expirationTime.Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}