package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type DepthexRespData struct {
	BatchSN           string              `json:"batchsn"`
	ExpressNo         string              `json:"expressno"`
	ExpressName       string              `json:"expressname"`
	Recipient         string              `json:"recipient"`
	Recipienttel      string              `json:"recipienttel"`
	RecipientAddress  string              `json:"recipientaddress"`
	RecipientCardNo   string              `json:"recipientcardno"`
	RecipientShopInfo string              `json:"recipientshopinfo"`
	Status            string              `json:"status"`
	StatusTime        int                 `json:"statustime"`
	ChinaExpress      []map[string]string `json:"chinaexpress"`
}

type DepthexResp struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Time string           `json:"time"`
	Data *DepthexRespData `json:"data"`
}

type OrderQueryTimer struct {
	db     *MysqlHandle
	ticker *time.Ticker
}

func NewOrderQueryTimer(db *MysqlHandle) *OrderQueryTimer {
	return &OrderQueryTimer{
		db:     db,
		ticker: time.NewTicker(time.Minute * 30),
	}
}

func (q *OrderQueryTimer) IfNeedUpdate(orderID string, statusTime string) bool {
	rows := q.db.ExecQuery("select statustime from logist.logistic_order where third_party_order_id = ?", orderID)
	var oldStatusTime string
	for rows.Next() {
		rows.Scan(&oldStatusTime)
	}
	if oldStatusTime == statusTime {
		return false
	}
	return true
}

func (q *OrderQueryTimer) parseDepthexResponse(data []byte, orderID string) {
	resp := &DepthexResp{}
	if err := json.Unmarshal(data, resp); err != nil {
		LogError("parse depthex failed, %s  %s for orderid %s ", err.Error(), string(data), orderID)
		result := make(map[string]interface{})
		json.Unmarshal(data, &result)
		if result["data"] == nil {
			LogError("data is nil")
			return
		}
		strResult := result["data"].(string)
		LogInfo("parse result, %#v", result)
		q.db.Exec("delete from logist.logistic_order_status where third_party_order_id = ? ", orderID)
		sql2 := "insert into logist.logistic_order_status(third_party_order_id, order_status_desc, order_status_time) values (?,?,?)"
		err3 := q.db.Exec(sql2, orderID, strResult, time.Now().Format("2006-01-02 15:04:05"))
		if err3 != nil {
			LogError("insert status to db failed %s, for %s", err3.Error(), orderID)
		}
		return
	}
	LogInfo("RESP: %s for order id  %s", string(data), orderID)
	if resp.Code == 1 { //success
		var lastStatus string
		var isFinished string

		if len(resp.Data.ChinaExpress) > 0 {
			lastStatus = resp.Data.ChinaExpress[0]["step"]
		}
		if strings.Contains(lastStatus, "已签收") {
			isFinished = "Y"
		} else {
			isFinished = "N"
		}
		statusTime := fmt.Sprintf("%d", resp.Data.StatusTime)
		if !q.IfNeedUpdate(orderID, statusTime) {
			LogInfo("don't need to update for order %s", orderID)
			return
		}
		sql := "update logist.logistic_order set batchsn = '%s' , expressno = '%s' , expressname = '%s', recipient = '%s', " +
			" recipienttel = '%s', recipientaddress = '%s', recipientcardno = '%s', recipientshopinfo = '%s', last_status = '%s', " +
			" statustime = '%s', is_finished = '%s' , create_time = now() where third_party_order_id = '%s'"
		sql = fmt.Sprintf(sql, resp.Data.BatchSN, resp.Data.ExpressNo, resp.Data.ExpressName,
			resp.Data.Recipient, resp.Data.Recipienttel, resp.Data.RecipientAddress, resp.Data.RecipientCardNo,
			resp.Data.RecipientShopInfo, lastStatus, statusTime, isFinished, orderID)

		LogInfo("sql %s", sql)
		err2 := q.db.Exec(sql)
		if err2 != nil {
			LogError("update logistic_order failed, %s for %s", err2.Error(), orderID)
			return
		}
		if len(resp.Data.ChinaExpress) > 0 {
			q.db.Exec("delete from logist.logistic_order_status where third_party_order_id = ? ", orderID)
		}
		for _, status := range resp.Data.ChinaExpress {
			sql2 := "insert into logist.logistic_order_status(third_party_order_id, order_status_desc, order_status_time) values (?,?,?)"
			err3 := q.db.Exec(sql2, orderID, status["step"], status["time"])
			if err3 != nil {
				LogError("insert status to db failed %s, for %s", err3.Error(), orderID)
			}
		}
	} else {
		LogError("query depthex response show error code , %d", resp.Code)
	}
}

func (q *OrderQueryTimer) doQuery() {
	querySql := "select a.third_party_order_id, b.third_party_api, b.third_party_name, a.is_finished " +
		"from logist.logistic_order a " +
		"left join logist.logistic_thirdparty b on a.third_party_id = b.id where a.is_finished = 'N' order by a.id desc"
	rows := q.db.ExecQuery(querySql)

	for rows.Next() {
		var thirdpartyOrderID, thirdpartyApi, thirdpartyName, isFinished string
		rows.Scan(&thirdpartyOrderID, &thirdpartyApi, &thirdpartyName, &isFinished)
		url := thirdpartyApi + thirdpartyOrderID
		go func() {
			LogInfo("query url : %s", url)
			resp, err := http.Get(url)
			if err != nil {
				LogError("query order status failed, %s %s", err.Error(), url)
				return
			}
			b, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				LogError("read response failed, %s %s", err2.Error(), url)
				return
			}
			if thirdpartyName == "depthex" {
				q.parseDepthexResponse(b, thirdpartyOrderID)

			} else if thirdpartyName == "" {
				LogError("thirdparty is empty")
			} else {
				LogError("not supported thirdparty %s", thirdpartyName)
			}
		}()

	}
}

func (q *OrderQueryTimer) Start() {
	for {
		select {
		case <-q.ticker.C:
			q.doQuery()
		}
	}
}
