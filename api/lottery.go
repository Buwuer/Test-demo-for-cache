package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"time"
	"webcache/db"
)

type Prize struct {
	source string // 0 代表正常礼品，1 代表未抽中奖品
	id     int16
	name   string
}

func (p Prize) insetLotteryRecord(memberId int32, activityId int16) error {
	stmtIns, err := db.MysqlClient.Prepare(
		"INSERT INTO LotteryRecord (member_id, activity_id, activityprize_id) VALUES( ?, ?, ? )")
	if err != nil {
		return errors.New("重新尝试")
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(memberId, activityId, p.id)
	if err != nil {
		return errors.New("重新尝试")
	}
	return nil
}

func (p Prize) givePrize(Id int32) error {
	stmtIns, err := db.MysqlClient.Prepare(
		"INSERT INTO MemberCoupon (member_id, coupon_name) VALUES( ?, ? )")
	if err != nil {
		return errors.New("重新尝试")
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(Id, p.name)
	if err != nil {
		return errors.New("重新尝试")
	}
	return nil
}

func (p Prize) reduceStock() error {
	stmtIns, err := db.MysqlClient.Prepare("UPDATE ActivityPrizes SET amount=amount - 1 where id = ?")
	if err != nil {
		return errors.New("重新尝试1")
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(p.id)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("重新尝试2")
	}
	return nil
}

type Lottery struct {
	UserId     int32 `json:"user_id" binding:"required"`
	ActivityId int16 `json:"activity_id" binding:"required"`
}

func LotteryApi(c *gin.Context) {
	// 参数校验
	//data, _ := ioutil.ReadAll(c.Request.Body)
	//fmt.Printf("ctx.Request.body: %v\n", string(data))
	var json Lottery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("validate", err.Error())
		return
	}
	// 会员身份校验，对UserId做了查询校验。其实可以考虑放到jwt中去，省一次查询
	if err := checkMember(json.UserId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 抽奖合法性校验，对ActivityId做查询校验，可以做到缓存中，省略一次查询
	if err := checkActivity(json.ActivityId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 抽取奖品
	prize, err := getPrize(json.ActivityId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("getPrize")
		return
	}
	// 扣减库存
	err = prize.reduceStock()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("reduceStock")
		return
	}
	// 抽奖记录写入，消息队列实现
	err = prize.insetLotteryRecord(json.UserId, json.ActivityId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("insetLotteryRecord")
		return
	}
	// 奖品数据写入，也可以做成服务单独调用, 也可以是消息队列实现
	err = prize.givePrize(json.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("givePrize")
		return
	}
	// 返回中奖信息
	c.JSON(200, gin.H{
		"prize":  prize.name,
		"source": prize.source,
	})
}

func getPrize(Id int16) (Prize, error) {
	Prizes := getPrizes(Id)
	PrizesLen := len(Prizes)

	if PrizesLen <= 1 {
		return Prize{}, errors.New("奖品全部抽完")
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(PrizesLen)
	return Prizes[n], nil
}

func getPrizes(Id int16) []Prize {
	rows, err := db.MysqlClient.Query("select id, name, source from ActivityPrizes where activity_id = ? and amount > 0", Id)
	if err != nil {
		log.Fatal(err)
		return []Prize{}
	}
	defer rows.Close()

	var list []Prize
	for rows.Next() {
		var prize Prize
		err := rows.Scan(&prize.id, &prize.name, &prize.source)
		if err != nil {
			log.Fatal(err)
			return []Prize{}
		}
		list = append(list, prize)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		return []Prize{}
	}

	return list
}

func checkMember(Id int32) error {
	stmtOut, err := db.MysqlClient.Prepare("SELECT name FROM Member WHERE id = ?")
	if err != nil {
		return errors.New("系统连接失败")
	}
	defer stmtOut.Close()

	var name string
	err = stmtOut.QueryRow(Id).Scan(&name)
	if err != nil {
		return errors.New("会员状态不正常")
	}
	return nil
}

func checkActivity(Id int16) error {
	stmtOut, err := db.MysqlClient.Prepare("SELECT * FROM Activity WHERE id = ?")
	if err != nil {
		return errors.New("系统连接失败")
	}
	defer stmtOut.Close()

	var activity struct {
		id    int
		state bool
	}
	err = stmtOut.QueryRow(Id).Scan(&activity.id, &activity.state)
	if err != nil {
		return errors.New("活动状态不正常")
	}
	return nil
}
