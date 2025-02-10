package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const (
	retryAttempts = 3
)

type PgStorage struct {
	db  *sql.DB
	ctx context.Context
}

func New(ctx context.Context, dsn string) (*PgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = applyMigrations(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PgStorage{db: db, ctx: ctx}, nil
}

func (s *PgStorage) UpdateGaugeMetric(name string, value float64) error {
	query := `INSERT INTO metric (id, type, value, delta) VALUES ($1, $2, $3, $4)
              ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value`

	var err error
	_ = retry.Do(
		func() error {
			_, err = s.db.ExecContext(s.ctx, query, name, model.MetricTypeGauge, value, nil)
			if isRetryableError(err) {
				return err
			}
			return nil
		},
		retry.Attempts(retryAttempts),
		retry.DelayType(retryDelayType),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Error(fmt.Sprintf("retry #%d to update gauge metric", n))
		}),
	)

	return err
}

func (s *PgStorage) UpdateCounterMetric(name string, value int64) error {
	query := `INSERT INTO metric (id, type, value, delta) VALUES ($1, $2, $3, $4)
              ON CONFLICT (id) DO UPDATE SET delta = metric.delta + EXCLUDED.delta`

	var err error
	_ = retry.Do(
		func() error {
			_, err := s.db.ExecContext(s.ctx, query, name, model.MetricTypeCounter, nil, value)
			if isRetryableError(err) {
				return err
			}
			return nil
		},
		retry.Attempts(retryAttempts),
		retry.DelayType(retryDelayType),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Error(fmt.Sprintf("retry #%d to update counter metric", n))
		}),
	)

	return err
}

func (s *PgStorage) GetGaugeMetric(metricName string) (float64, error) {
	var value sql.NullFloat64
	query := `SELECT value FROM metric WHERE type = 'gauge' and id = $1`
	row := s.db.QueryRowContext(s.ctx, query, metricName)
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	if !value.Valid {
		return 0, sql.ErrNoRows
	}

	return value.Float64, nil
}

func (s *PgStorage) GetCounterMetric(metricName string) (int64, error) {
	var delta sql.NullInt64
	query := `SELECT delta FROM metric WHERE type = 'counter' and id = $1`
	row := s.db.QueryRowContext(s.ctx, query, metricName)
	err := row.Scan(&delta)
	if err != nil {
		return 0, err
	}

	if !delta.Valid {
		return 0, sql.ErrNoRows
	}

	return delta.Int64, nil
}

func (s *PgStorage) GetAllMetrics() ([]model.Metrics, error) {
	var metrics []model.Metrics
	query := `SELECT id, type, value, delta FROM metric`
	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *PgStorage) SetAllMetrics(metrics []model.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO metric (id, type, value, delta) VALUES($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET delta = metric.delta + EXCLUDED.delta, value = EXCLUDED.value
	`

	stmt, err := tx.PrepareContext(s.ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range metrics {
		_, err := stmt.ExecContext(s.ctx, m.ID, m.MType, m.Value, m.Delta)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PgStorage) Ping() error {
	return s.db.Ping()
}

func applyMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

func isRetryableError(err error) bool {
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		return true
	}

	return false
}

func retryDelayType(n uint, err error, config *retry.Config) time.Duration {
	switch n {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	default:
		return 5 * time.Second
	}
}
