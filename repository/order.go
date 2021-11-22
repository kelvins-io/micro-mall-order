package repository

import (
	"time"

	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/kelvins-io/kelvins"
	"xorm.io/xorm"
)

func CreateOrder(tx *xorm.Session, models []mysql.Order) (err error) {
	_, err = tx.Table(mysql.TableOrder).Insert(models)
	return
}

func GetOrderExist(txCode string) (bool, error) {
	var model mysql.Order
	var err error
	_, err = kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select("tx_code,order_code").Where("tx_code = ?", txCode).Get(&model)
	if err != nil {
		return false, err
	}
	if model.TxCode != "" && model.OrderCode != "" {
		return true, nil
	}
	return false, nil
}

func FindOrderListByOrderCode(sqlSelect string, orderCode []string) ([]mysql.Order, error) {
	var result = make([]mysql.Order, 0)
	var err error
	err = kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select(sqlSelect).In("order_code", orderCode).Find(&result)
	return result, err
}

func OrderShopRank(sqlSelect string, where map[string]interface{}, groupBy string, inOrder []string, startTime, endTime time.Time, pageSize, pageNum int32) ([]mysql.OrderShopRank, error) {
	var result = make([]mysql.OrderShopRank, 0)
	var err error
	session := kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select(sqlSelect).Where(where)
	if startTime.Year() > 1999 {
		session = session.Where("create_time >= ?", startTime)
	}
	if endTime.Year() > 1999 {
		session = session.Where("create_time < ?", endTime)
	}
	if pageSize > 0 && pageNum > 0 {
		session = session.Limit(int(pageSize), int((pageNum-1)*pageSize))
	}
	err = session.GroupBy(groupBy).Desc(inOrder...).Find(&result)
	return result, err
}

func FindOrderListByTime(sqlSelect string, where interface{}, startTime, endTime time.Time, pageSize, pageNum int32) ([]mysql.Order, int64, error) {
	var result = make([]mysql.Order, 0)
	var err error
	var total int64
	session := kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select(sqlSelect)
	if where != nil {
		session = session.Where(where)
	}
	if startTime.Year() > 1999 {
		session = session.Where("create_time >= ?", startTime)
	}
	if endTime.Year() > 1999 {
		session = session.Where("create_time < ?", endTime)
	}
	if pageSize > 0 && pageNum > 0 {
		session = session.Limit(int(pageSize), int((pageNum-1)*pageSize))
	}
	session = session.Desc("create_time")
	err = session.Find(&result)
	total = int64(len(result))
	return result, total, err
}

func GetOrderList(selSelect string, where interface{}) ([]mysql.Order, error) {
	var result = make([]mysql.Order, 0)
	var err error
	err = kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select(selSelect).Where(where).Find(&result)
	return result, err
}

func UpdateOrder(query, maps interface{}) (int64, error) {
	return kelvins.XORM_DBEngine.Table(mysql.TableOrder).Where(query).Update(maps)
}

func UpdateOrderByTx(tx *xorm.Session, query, maps interface{}) (int64, error) {
	return tx.Table(mysql.TableOrder).Where(query).Update(maps)
}
