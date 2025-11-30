package storage

import (
	"database/sql"
	"log/slog"

	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/storage/mysql"
)

func GetMysql(name string) *sql.DB {
	return mysql.GetMysql(name)
}

func InitDB() {
	cnf := internalCnf.Get().MySQL
	mySQLs := []*mysql.MySQL{
		{
			DBName:   cnf.EventDBM.DBName,
			Account:  cnf.EventDBM.Account,
			Password: cnf.EventDBM.Password,
			IP:       cnf.EventDBM.IP,
			Port:     cnf.EventDBM.Port,
		},
		{
			DBName:   cnf.EventDBS.DBName,
			Account:  cnf.EventDBS.Account,
			Password: cnf.EventDBS.Password,
			IP:       cnf.EventDBS.IP,
			Port:     cnf.EventDBS.Port,
		},
	}

	for _, v := range mySQLs {
		if err := mysql.InitMysql(v); err != nil {
			slog.Error("failed to init mysql", slog.Any("err", err))
			panic(err)
		}
	}
}
