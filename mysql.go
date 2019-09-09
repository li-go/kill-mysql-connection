package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type DBConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewDBClient(config DBConfig, network string) (*DBClient, error) {
	log.Printf("Connecting MySQL server - %s@%s:%d ...", config.Username, config.Host, config.Port)
	ds := fmt.Sprintf("%s:%s@%s(%s:%d)/", config.Username, config.Password, network, config.Host, config.Port)
	db, err := sqlx.Open("mysql", ds)
	if err != nil {
		return nil, fmt.Errorf("fail to connect MySQL: %w", err)
	}
	return &DBClient{db: db}, nil
}

type DBClient struct {
	db *sqlx.DB
}

func (c *DBClient) Close() error {
	return c.db.Close()
}

func (c *DBClient) GetProcesslist() ([]Process, error) {
	var processlist []Process
	err := c.db.Select(&processlist, `SELECT * FROM information_schema.PROCESSLIST ORDER BY TIME DESC`)
	if err != nil {
		return nil, fmt.Errorf("fail to select PROCESSLIST: %w", err)
	}
	return processlist, nil
}

func (c *DBClient) KillProcess(processID int) error {
	_, err := c.db.Exec(`KILL ?`, processID)
	if err != nil {
		return fmt.Errorf("fail to KILL process - %d: %w", processID, err)
	}
	return nil
}

type Process struct {
	ID      int            `db:"ID"`
	User    string         `db:"USER"`
	Host    string         `db:"HOST"`
	DB      sql.NullString `db:"DB"`
	Command string         `db:"COMMAND"`
	Time    int            `db:"TIME"`
	State   sql.NullString `db:"STATE"`
	Info    sql.NullString `db:"INFO"`
}

func (p Process) Keys() []string {
	return []string{"ID", "USER", "HOST", "DB", "COMMAND", "TIME", "STATE", "INFO"}
}

func (p Process) Values() []string {
	return []string{
		strconv.Itoa(p.ID), p.User, p.Host, p.DB.String, p.Command, strconv.Itoa(p.Time), p.State.String, p.Info.String,
	}
}

func PrintProcesslist(processlist []Process) {
	maxLens := make([]int, 8)
	for _, p := range processlist {
		for i, v := range p.Values() {
			if len(v) > maxLens[i] {
				maxLens[i] = len(v)
			}
		}
	}

	fmt.Printf("NO.\t")
	for i, k := range (Process{}).Keys() {
		fmt.Printf(fmt.Sprintf("%%-%ds", maxLens[i]+4), k)
	}
	fmt.Println()

	for pi, p := range processlist {
		fmt.Printf("%d.\t", pi+1)
		for i, v := range p.Values() {
			fmt.Printf(fmt.Sprintf("%%-%ds", maxLens[i]+4), v)
		}
		fmt.Println()
	}
}
