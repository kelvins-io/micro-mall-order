package repository

import (
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"xorm.io/xorm"
)

func CreateOrderSceneShop(tx *xorm.Session, models []mysql.OrderSceneShop) (err error) {
	_, err = tx.Table(mysql.TableOrderSceneShop).Insert(models)
	return
}
