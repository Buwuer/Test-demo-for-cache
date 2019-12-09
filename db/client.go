package db

import (
	"database/sql"
	//"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

var MysqlClient *sql.DB

func init() {
	MysqlClient, _ = sql.Open("mysql", "root:mariadb@tcp(120.27.245.213:3306)/wecache?charset=utf8mb4")

	//RedisClient  := redis.NewClient(&redis.Options{
	//	Addr:     "120.27.245.213:23306",
	//	Password: "", // no password set
	//	DB:       0,  // use default DB
	//})

	/*
		val, err := client.Get(key).Result()
			if err != nil {
				if err == redis.Nil {
					fmt.Println(key, "does not exist")
				} else {
					panic(err)
				}

			}
	*/
}
