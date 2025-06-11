package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// CheckInHandler performs a check-in or check-out based on the token and Telegram ID.
func (e *Env) CheckInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	tgStr := r.URL.Query().Get("tg_id")
	if tgStr == "" {
		http.Error(w, "missing tg_id", http.StatusBadRequest)
		return
	}
	tgID, err := strconv.ParseInt(tgStr, 10, 64)
	if err != nil {
		http.Error(w, "bad tg_id", http.StatusBadRequest)
		return
	}

	if _, err := e.App.Redis.GetDel(ctx, token).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			http.Error(w, "invalid or expired token", http.StatusBadRequest)
		} else {
			http.Error(w, "redis error", http.StatusInternalServerError)
		}
		return
	}

	tx, err := e.App.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var empID int
	err = tx.QueryRow(ctx, "select id from employees where telegram_id=$1", tgID).Scan(&empID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "insert into employees(telegram_id) values($1) returning id", tgID).Scan(&empID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	var shiftID int
	var start time.Time
	err = tx.QueryRow(ctx, "select id, start_time from shifts where employee_id=$1 and end_time is null", empID).Scan(&shiftID, &start)
	if errors.Is(err, pgx.ErrNoRows) {
		// perform check-in
		_, err = tx.Exec(ctx, "insert into shifts(employee_id,start_time,ip,user_agent) values($1,$2,$3,$4)", empID, time.Now(), r.RemoteAddr, r.UserAgent())
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(ctx); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "checked in")
		return
	} else if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// existing open shift -> check out
	_, err = tx.Exec(ctx, "update shifts set end_time=$1 where id=$2", time.Now(), shiftID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	duration := time.Since(start).Truncate(time.Second)
	fmt.Fprintf(w, "checked out, duration %s", duration)
}
