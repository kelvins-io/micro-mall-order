package repository

import (
	"time"

	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/kelvins-io/kelvins"
	"xorm.io/xorm"
)

func CreateOrderSku(tx *xorm.Session, models []mysql.OrderSku) (err error) {
	_, err = tx.Table(mysql.TableOrderSku).Insert(models)
	return
}

func GetOrderSkuListByOrderCode(sqlSelect string, orderCode []string) ([]mysql.OrderSku, error) {
	var result = make([]mysql.OrderSku, 0)
	var err error
	err = kelvins.XORM_DBEngine.Table(mysql.TableOrderSku).Select(sqlSelect).In("order_code", orderCode).Find(&result)
	return result, err
}

func FindOrderSkuByOrderCode(sqlSelect string, orderCode []string) ([]mysql.OrderSku, error) {
	var result = make([]mysql.OrderSku, 0)
	var err error
	err = kelvins.XORM_DBEngine.Table(mysql.TableOrderSku).Select(sqlSelect).In("order_code", orderCode).Find(&result)
	return result, err
}

func OrderSkuRank(sqlSelect string, where map[string]interface{}, groupBy string, inOrder []string, startTime, endTime time.Time, pageSize, pageNum int32) ([]mysql.OrderSkuRank, error) {
	var result = make([]mysql.OrderSkuRank, 0)
	var err error
	session := kelvins.XORM_DBEngine.Table(mysql.TableOrderSku).Select(sqlSelect).Where(where)
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
