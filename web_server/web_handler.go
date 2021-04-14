package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type WebHandlers struct {
	db *MysqlHandle
}

func NewWebHandlers(db *MysqlHandle) *WebHandlers {
	return &WebHandlers{
		db: db,
	}
}

type QueryStatusReq struct {
	OrderID []string `json:"order_id"`
}

type OrderStatusObj struct {
	OrderID    string `json:"order_id"`
	Desc       string `json:"desc"`
	UpdateTime string `json:"update_time"`
	ExpressNo  string `json:"expressno"`
}

type QueryStatusResp struct {
	Code   int               `json:"code"`
	Status []*OrderStatusObj `json:"status"`
}

//request: { "order_id" : ["123123123123123123", ...]}
/*response :
{
	"code" : 0 // 0 ok, or error
	"status" : [
		{
			"order_id" : "123123123123123123",
			"status" : "xxxxxx",
			"desc": "adfsadf",
			"update_time" : "2021-12-21 11:23:34",
			"expressno" : "xxxxxxx"
		},
		...
	]
}
*/
func (h *WebHandlers) HandleQueryOrderStatus(resp http.ResponseWriter, req *http.Request) {
	LogInfo("handle query order status request")
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		ErrorResponse(resp, -1)
		return
	}
	var reqData QueryStatusReq
	if err2 := json.Unmarshal(b, &reqData); err2 != nil {
		ErrorResponse(resp, -2)
		return
	}
	inSql := "( "
	for _, orderID := range reqData.OrderID {
		inSql += "\"" + orderID + "\","
	}
	inSql = inSql[0 : len(inSql)-1]
	inSql += " )"
	sql := fmt.Sprintf("SELECT a.our_order_id, b.order_status_desc, b.order_status_time , ifnull(a.expressno, '') as expressno, "+
		"ifnull(a.expressname, '') as expressname, ifnull(a.recipient, '') as recipient, ifnull(a.recipienttel, '') as recipienttel, "+
		"ifnull(a.recipientaddress,'') as recipientaddress, ifnull(a.batchsn, '') as batchsn "+
		"FROM logist.logistic_order a  left join logist.logistic_order_status b on a.third_party_order_id = b.third_party_order_id "+
		"where a.our_order_id in %s order by a.our_order_id, b.order_status_time desc", inSql)
	rows := h.db.ExecQuery(sql)
	var orderID, orderStatusDesc, updateTime string
	var expressNO, expressName, recipient, recipientTel, recipientAddress, batchsn string
	orderResp := &QueryStatusResp{
		Status: []*OrderStatusObj{},
	}
	orderResp.Code = 0
	for rows.Next() {
		rows.Scan(&orderID, &orderStatusDesc, &updateTime, &expressNO, &expressName, &recipient, &recipientTel, &recipientAddress, &batchsn)
		if len(batchsn) == 0 {
			obj := &OrderStatusObj{
				OrderID:    orderID,
				Desc:       "待发货",
				UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
				ExpressNo:  expressNO,
			}
			orderResp.Status = append(orderResp.Status, obj)
		} else {
			obj := &OrderStatusObj{
				OrderID:    orderID,
				Desc:       orderStatusDesc,
				UpdateTime: updateTime,
				ExpressNo:  expressNO,
			}
			orderResp.Status = append(orderResp.Status, obj)
		}

	}
	o, _ := json.Marshal(orderResp)
	resp.Write(o)
}
