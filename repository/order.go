package repository

import (
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/kelvins-io/kelvins"
	"xorm.io/xorm"
)

func CreateOrder(tx *xorm.Session, models []mysql.Order) (err error) {
	_, err = tx.Table(mysql.TableOrder).Insert(models)
	return
}

func GetOrderExist(txCode string) (*mysql.Order, error) {
	var model mysql.Order
	var err error
	_, err = kelvins.XORM_DBEngine.Table(mysql.TableOrder).Select("tx_code,order_code").Where("tx_code = ?", txCode).Get(&model)
	return &model, err
}

func GetOrderListByTxCode(txCode string) ([]mysql.Order, error) {
	var result = make([]mysql.Order, 0)
	var err error
	err = kelvins.XORM_DBEngine.Table(mysql.TableOrder).Where("tx_code = ?", txCode).Find(&result)
	return result, err
}

func UpdateOrder(query, maps interface{}) (int64, error) {
	return kelvins.XORM_DBEngine.Table(mysql.TableOrder).Where(query).Update(maps)
}

func UpdateOrderByTx(tx *xorm.Session, query, maps interface{}) (int64, error) {
	return tx.Table(mysql.TableOrder).Where(query).Update(maps)
}
