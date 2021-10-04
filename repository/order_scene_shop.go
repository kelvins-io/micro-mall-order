package repository

import (
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/kelvins-io/kelvins"
	"xorm.io/xorm"
)

func CreateOrderSceneShop(tx *xorm.Session, models []mysql.OrderSceneShop) (err error) {
	_, err = tx.Table(mysql.TableOrderSceneShop).Insert(models)
	return
}

func FindOrderSceneShop(sqlSelect string, orderCode []string) ([]mysql.OrderSceneShop, error) {
	var result = make([]mysql.OrderSceneShop, 0)
	err := kelvins.XORM_DBEngine.Table(mysql.TableOrderSceneShop).Select(sqlSelect).In("order_code", orderCode).Find(&result)
	return result, err
}
