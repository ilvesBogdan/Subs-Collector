package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"subs-collector/internal/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *model.Subscription) (int, error)
	GetByID(ctx context.Context, id int) (*model.Subscription, error)
	Update(ctx context.Context, id int, s *model.Subscription) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, userID, serviceName string) ([]model.Subscription, error)
	SumTotal(ctx context.Context, from, to time.Time, userID, serviceName string) (int, error)
}

type subscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepository{pool: pool}
}

func (r *subscriptionRepository) Create(ctx context.Context, s *model.Subscription) (int, error) {
	serviceID, err := r.ensureService(ctx, s.ServiceName)
	if err != nil {
		return 0, err
	}

	const sql = `INSERT INTO user_subscriptions (
	               service_id, price, user_id, start_date, end_date
	           ) VALUES ($1, $2, $3::uuid, $4, $5) RETURNING id`

	var id int
	err = r.pool.QueryRow(ctx, sql, serviceID, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&id)
	return id, err
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id int) (*model.Subscription, error) {
	const sql = `SELECT us.id, sv.name AS service_name, us.price, us.user_id::text, us.start_date, us.end_date
	           FROM user_subscriptions us
	           JOIN services sv ON sv.id = us.service_id
	           WHERE us.id=$1`

	row := r.pool.QueryRow(ctx, sql, id)
	var m model.Subscription
	err := row.Scan(&m.ID, &m.ServiceName, &m.Price, &m.UserID, &m.StartDate, &m.EndDate)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, id int, s *model.Subscription) error {
	serviceID, err := r.ensureService(ctx, s.ServiceName)
	if err != nil {
		return err
	}

	const sql = `UPDATE user_subscriptions 
	               SET service_id=$1, price=$2, user_id=$3::uuid, start_date=$4, end_date=$5, updated_at=$6 WHERE id=$7`
	ct, err := r.pool.Exec(ctx, sql, serviceID, s.Price, s.UserID, s.StartDate, s.EndDate, time.Now().UTC(), id)

	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}

	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id int) error {
	const sql = `DELETE FROM user_subscriptions WHERE id=$1`
	ct, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}

	return nil
}

func (r *subscriptionRepository) List(ctx context.Context, userID string, serviceName string) ([]model.Subscription, error) {
	sql := `SELECT us.id, sv.name AS service_name, us.price, us.user_id::text, us.start_date, us.end_date
	      FROM user_subscriptions us JOIN services sv ON sv.id = us.service_id WHERE 1=1`
	var args []interface{}
	idx := 1

	if userID != "" {
		sql += " AND us.user_id=$" + strconv.Itoa(idx) + "::uuid"
		args = append(args, userID)
		idx++
	}

	if serviceName != "" {
		sql += " AND sv.name=$" + strconv.Itoa(idx)
		args = append(args, serviceName)
		idx++
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.Subscription, 0)
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate); err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, rows.Err()
}

// SumTotal считает суммарную стоимость за каждый месяц периода [from..to] включительно,
// учитывая только те месяцы, в которых подписка активна. Если end_date NULL — бесконечная.
func (r *subscriptionRepository) SumTotal(ctx context.Context, from, to time.Time, userID, serviceName string) (int, error) {
	sql := `WITH months AS (
	            SELECT
	                generate_series(date_trunc('month', $1::timestamptz),
	                date_trunc('month', $2::timestamptz), interval '1 month') AS m
	     )
	     SELECT COALESCE(SUM(us.price), 0) AS total
	     FROM months mo
	     JOIN user_subscriptions us
	       ON date_trunc('month', us.start_date) <= mo.m
	      AND (us.end_date IS NULL OR date_trunc('month', us.end_date) >= mo.m)
	     JOIN services sv ON sv.id = us.service_id
	     WHERE ($3 = '' OR us.user_id = $3::uuid)
	       AND ($4 = '' OR sv.name = $4)`
	var total int
	err := r.pool.QueryRow(ctx, sql, from, to, userID, serviceName).Scan(&total)
	return total, err
}

// ensureService возвращает id сервиса, создавая запись при необходимости
func (r *subscriptionRepository) ensureService(ctx context.Context, name string) (int, error) {
	const ins = `INSERT INTO services(name) VALUES ($1) ON CONFLICT (name) DO NOTHING`
	if _, err := r.pool.Exec(ctx, ins, name); err != nil {
		return 0, err
	}
	const sel = `SELECT id FROM services WHERE name=$1`
	var id int
	if err := r.pool.QueryRow(ctx, sel, name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
