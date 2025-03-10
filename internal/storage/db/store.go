package db

import (
	"context"
	"errors"
	"fmt"
	"httpserver/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPool(cfg Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	return pgxpool.New(context.Background(), connStr)
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

var ErrFilePathExists = errors.New("file path already exists")

func (s *Store) SaveFile(ctx context.Context, file *models.File) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx,
		`WITH ins AS (
            INSERT INTO files (filepath, size, metadata) 
            VALUES ($1, $2, $3)
            ON CONFLICT (filepath) DO NOTHING
            RETURNING id
        )
        SELECT COALESCE(
            (SELECT id FROM ins),
            (SELECT id FROM files WHERE filepath = $1)
        )`,
		file.Filepath,
		file.Size,
		file.Metadata,
	).Scan(&id)

	return id, err
}

func (s *Store) FileExists(ctx context.Context, filepath string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM files WHERE filepath = $1)",
		filepath).Scan(&exists)
	return exists, err
}

func (s *Store) GetDicomTag(ctx context.Context, id int64, tag string) (string, error) {
	var value string
	err := s.db.QueryRow(ctx,
		`SELECT metadata->$2 
         FROM files 
         WHERE id = $1`,
		id, tag,
	).Scan(&value)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("file or tag not found")
		}
		return "", err
	}

	return value, nil
}

func (s *Store) GetFileByID(ctx context.Context, id int64) (*models.File, error) {
	var file models.File
	err := s.db.QueryRow(ctx,
		"SELECT id, filepath, size, metadata FROM files WHERE id = $1",
		id,
	).Scan(&file.ID, &file.Filepath, &file.Size, &file.Metadata)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, err
	}

	return &file, nil
}
