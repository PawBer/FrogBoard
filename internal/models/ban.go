package models

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Ban struct {
	IP        net.IP
	Reason    string
	StartDate time.Time
	EndDate   time.Time
}

type BanModel struct {
	DbConn *goqu.Database
}

func (bm *BanModel) IsBanned(r *http.Request) (bool, Ban, error) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	convertedIP := net.ParseIP(host)

	query, params, _ := goqu.From("bans").Select("reason", "start_date", "end_date").Where(goqu.Ex{
		"ip": convertedIP.String(),
	}).ToSQL()

	var ban Ban
	err := bm.DbConn.QueryRow(query, params...).Scan(&ban.Reason, &ban.StartDate, &ban.EndDate)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, Ban{}, nil
	}
	if err != nil {
		return false, Ban{}, err
	}
	ban.IP = convertedIP

	return true, ban, nil
}

func (bm *BanModel) GetBan(ip net.IP) (Ban, error) {
	var ban Ban
	var ipStr string

	query, params, _ := goqu.From("bans").Select("ip", "reason", "start_date", "end_date").Where(goqu.Ex{
		"ip": ip.String(),
	}).ToSQL()

	err := bm.DbConn.QueryRow(query, params...).Scan(&ipStr, &ban.Reason, &ban.StartDate, &ban.EndDate)
	if err != nil {
		return Ban{}, err
	}

	ban.IP = net.ParseIP(ipStr)

	return ban, nil
}

func (bm *BanModel) GetBans(pageNumber, itemsPerPage uint) ([]Ban, error) {
	var bans []Ban

	query, params, _ := goqu.From("bans").Select("ip", "reason", "start_date", "end_date").Order(goqu.I("start_date").Desc()).Limit(itemsPerPage).Offset(pageNumber * itemsPerPage).ToSQL()

	rows, err := bm.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	var ban Ban
	var ip string
	for rows.Next() {
		err := rows.Scan(&ip, &ban.Reason, &ban.StartDate, &ban.EndDate)
		if err != nil {
			return nil, err
		}
		ban.IP = net.ParseIP(ip)

		bans = append(bans, ban)
	}

	return bans, nil
}

func (m *ThreadModel) GetBanCount() (uint, error) {
	query, params, _ := goqu.From("bans").Select(goqu.COUNT("*")).ToSQL()

	var count uint
	err := m.DbConn.QueryRow(query, params...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (bm *BanModel) BanUser(ip net.IP, endDate time.Time, reason string) error {
	startDate := time.Now().UTC()

	query, params, _ := goqu.Insert("bans").Rows(goqu.Record{
		"ip":         ip.String(),
		"reason":     reason,
		"start_date": startDate,
		"end_date":   endDate,
	}).ToSQL()

	_, err := bm.DbConn.Exec(query, params...)
	if err != nil {
		return err
	}

	return nil
}

func (bm *BanModel) UnbanUser(ip net.IP) error {
	query, params, _ := goqu.Delete("bans").Where(goqu.Ex{
		"ip": ip.String(),
	}).ToSQL()

	_, err := bm.DbConn.Exec(query, params...)
	if err != nil {
		return err
	}

	return nil
}
