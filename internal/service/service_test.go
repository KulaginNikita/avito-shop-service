package service_test

import (
	"database/sql"
	"regexp"
	"testing"
	"golang.org/x/crypto/bcrypt"
	"avito-shop/internal/config"
	"avito-shop/internal/repository"
	"avito-shop/internal/service"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Тесты AuthUser
// -----------------------------------------------------------------------------

func TestAuthUser_NewUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{JWTSecret: "test-secret"}
	svc := service.NewService(repo, cfg)

	username := "newuser"
	password := "secret123"

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password, coins FROM users WHERE username = $1`,
	)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	mock.
		ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO users (username, password, coins) VALUES ($1, $2, 1000) RETURNING id`,
		)).
		WithArgs(username, sqlmock.AnyArg()). 
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	token, err := svc.AuthUser(username, password)
	require.NoError(t, err)
	assert.NotEmpty(t, token, "JWT token should be returned")

	require.NoError(t, mock.ExpectationsWereMet())
}


func TestAuthUser_ExistingUser_OK(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := repository.NewRepository(db)
    cfg := &config.Config{JWTSecret: "test-secret"}
    svc := service.NewService(repo, cfg)

    username := "alice"
    password := "pwd"

    realHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id, username, password, coins FROM users WHERE username = $1`,
    )).
        WithArgs(username).
        WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
            AddRow(1, "alice", string(realHash), 1000))

    token, err := svc.AuthUser(username, password)
    require.NoError(t, err, "AuthUser should succeed with correct password")
    assert.NotEmpty(t, token)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthUser_ExistingUser_BadPassword(t *testing.T) {
	// Сценарий: пользователь существует, но пароль неверен => ErrInvalidPassword
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{JWTSecret: "test-secret"}
	svc := service.NewService(repo, cfg)

	username := "alice"
	password := "wrong-pass" 

	hashedPass := "$2a$10$IXpQW...someHashOfRealPassword...vZ8S9Eu"

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, username, password, coins FROM users WHERE username = $1`,
	)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
			AddRow(1, "alice", hashedPass, 1000))

	token, err := svc.AuthUser(username, password)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
	assert.Empty(t, token)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthUser_EmptyCredentials(t *testing.T) {
	// Пустой username/password => ErrInvalidPassword
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	token, err := svc.AuthUser("", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
	assert.Empty(t, token)
}

// -----------------------------------------------------------------------------
// Тесты SendCoin
// -----------------------------------------------------------------------------

func TestSendCoin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{JWTSecret: "test-secret"}
	svc := service.NewService(repo, cfg)

	fromUserID := 1
	toUsername := "bob"
	amount := 100

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE id = $1`)).
		WithArgs(fromUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
			AddRow(1, "alice", "somepass", 500))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE username = $1`)).
		WithArgs(toUsername).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
			AddRow(2, "bob", "passbob", 200))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET coins = $1 WHERE id = $2`)).
		WithArgs(400, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET coins = $1 WHERE id = $2`)).
		WithArgs(300, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO coin_transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)`)).
		WithArgs(1, 2, 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.SendCoin(fromUserID, toUsername, amount)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCoin_NotEnoughCoins(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := repository.NewRepository(db)
    cfg := &config.Config{}
    svc := service.NewService(repo, cfg)

    fromUserID := 1
    toUsername := "bob"
    amount := 1000

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id, username, password, coins FROM users WHERE id = $1`,
    )).
        WithArgs(fromUserID).
        WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
            AddRow(1, "alice", "somepass", 200))

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id, username, password, coins FROM users WHERE username = $1`,
    )).
        WithArgs(toUsername).
        WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
            AddRow(2, "bob", "passbob", 500))


    err = svc.SendCoin(fromUserID, toUsername, amount)
    assert.EqualError(t, err, "not enough coins")

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCoin_ReceiverNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	fromUserID := 1
	toUsername := "unknown"
	amount := 50

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE id = $1`)).
		WithArgs(fromUserID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
				AddRow(1, "alice", "somepass", 500),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE username = $1`)).
		WithArgs(toUsername).
		WillReturnError(sql.ErrNoRows)

	err = svc.SendCoin(fromUserID, toUsername, amount)
	assert.EqualError(t, err, "recipient not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

// -----------------------------------------------------------------------------
// Тесты BuyItem
// -----------------------------------------------------------------------------

func TestBuyItem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	userID := 10
	itemName := "t-shirt" // 80 монет

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
				AddRow(10, "alice", "somepass", 200),
		)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET coins = $1 WHERE id = $2`)).
		WithArgs(120, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO item_purchases (user_id, item_name, quantity) VALUES ($1, $2, $3)`)).
		WithArgs(10, "t-shirt", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO coin_transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)`)).
		WithArgs(10, nil, 80).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.BuyItem(userID, itemName)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItem_NotEnoughCoins(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	userID := 10
	itemName := "t-shirt" 

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "password", "coins"}).
				AddRow(10, "alice", "somepass", 50),
		)

	err = svc.BuyItem(userID, itemName)
	assert.EqualError(t, err, "not enough coins")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItem_UnknownItem(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	userID := 10
	itemName := "some-weird-item"

	err = svc.BuyItem(userID, itemName)
	assert.EqualError(t, err, "invalid item")
}


// -----------------------------------------------------------------------------
// Тест GetInfo
// -----------------------------------------------------------------------------


func TestGetInfo_NoUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewRepository(db)
	cfg := &config.Config{}
	svc := service.NewService(repo, cfg)

	userID := 999

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, username, password, coins FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	info, err := svc.GetInfo(userID)
	assert.Nil(t, info)
	assert.EqualError(t, err, "user not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}
