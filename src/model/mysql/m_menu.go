package mysql

import (
	"context"
	"fmt"
	"gin-admin/src/model"
	"gin-admin/src/schema"
	"gin-admin/src/service/mysql"
	"time"

	"github.com/facebookgo/inject"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Menu 菜单管理
type Menu struct {
	DB     *mysql.DB
	Common *Common
}

// Init 初始化
func (a *Menu) Init(g *inject.Graph, db *mysql.DB, c *Common) *Menu {
	a.DB = db
	a.Common = c

	g.Provide(&inject.Object{Value: model.IMenu(a), Name: "IMenu"})

	db.CreateTableIfNotExists(schema.Menu{}, a.TableName())
	db.CreateTableIndex(a.TableName(), "idx_record_id", true, "record_id")
	db.CreateTableIndex(a.TableName(), "idx_name", false, "name")
	db.CreateTableIndex(a.TableName(), "idx_type", false, "type")
	db.CreateTableIndex(a.TableName(), "idx_parent_id", false, "parent_id")
	db.CreateTableIndex(a.TableName(), "idx_status", false, "status")
	db.CreateTableIndex(a.TableName(), "idx_deleted", false, "deleted")

	return a
}

// TableName 表名
func (a *Menu) TableName() string {
	return fmt.Sprintf("%s_%s", viper.GetString("mysql_table_prefix"), "menu")
}

// QueryPage 查询分页数据
func (a *Menu) QueryPage(ctx context.Context, param schema.MenuQueryParam, pageIndex, pageSize uint) (int64, []*schema.MenuQueryResult, error) {
	var (
		where = "WHERE deleted=0"
		args  []interface{}
	)

	if v := param.Name; v != "" {
		where = fmt.Sprintf("%s AND name LIKE ?", where)
		args = append(args, "%"+v+"%")
	}
	if v := param.ParentID; v != "" {
		where = fmt.Sprintf("%s AND parent_id=?", where)
		args = append(args, v)
	}
	if v := param.Status; v > 0 {
		where = fmt.Sprintf("%s AND status=?", where)
		args = append(args, v)
	}
	if v := param.Type; v > 0 {
		where = fmt.Sprintf("%s AND type=?", where)
		args = append(args, v)
	}

	count, err := a.DB.SelectInt(fmt.Sprintf("SELECT COUNT(*) FROM %s %s", a.TableName(), where), args...)
	if err != nil {
		return 0, nil, errors.Wrap(err, "查询分页数据发生错误")
	} else if count == 0 {
		return 0, nil, nil
	}

	var items []*schema.MenuQueryResult
	fields := "id,record_id,code,name,icon,uri,type,sequence,status"
	_, err = a.DB.Select(&items, fmt.Sprintf("SELECT %s FROM %s %s ORDER BY type,sequence,id LIMIT %d,%d", fields, a.TableName(), where, (pageIndex-1)*pageSize, pageSize), args...)
	if err != nil {
		return 0, nil, errors.Wrap(err, "查询分页数据发生错误")
	}

	return count, items, nil
}

// QuerySelect 查询选择数据
func (a *Menu) QuerySelect(ctx context.Context, param schema.MenuSelectQueryParam) ([]*schema.MenuSelectQueryResult, error) {
	var (
		where = "WHERE deleted=0"
		args  []interface{}
	)

	if v := param.Name; v != "" {
		where = fmt.Sprintf("%s AND name LIKE ?", where)
		args = append(args, "%"+v+"%")
	}
	if v := param.Status; v > 0 {
		where = fmt.Sprintf("%s AND status=?", where)
		args = append(args, v)
	}

	var items []*schema.MenuSelectQueryResult
	fields := "record_id,name,level_code,parent_id"
	_, err := a.DB.Select(&items, fmt.Sprintf("SELECT %s FROM %s %s ORDER BY sequence,id", fields, a.TableName(), where), args...)
	if err != nil {
		return nil, errors.Wrap(err, "查询选择数据发生错误")
	}

	return items, nil
}

// Get 查询指定数据
func (a *Menu) Get(ctx context.Context, recordID string) (*schema.Menu, error) {
	var item schema.Menu
	fields := "id,record_id,code,name,type,sequence,icon,uri,level_code,parent_id,status,creator,created,deleted"

	err := a.DB.SelectOne(&item, fmt.Sprintf("SELECT %s FROM %s WHERE deleted=0 AND record_id=?", fields, a.TableName()), recordID)
	if err != nil {
		return nil, errors.Wrap(err, "查询指定数据发生错误")
	}
	return &item, nil
}

// Create 创建数据
func (a *Menu) Create(ctx context.Context, item *schema.Menu) error {
	err := a.DB.Insert(item)
	if err != nil {
		return errors.Wrap(err, "创建数据发生错误")
	}
	return nil
}

// Update 更新数据
func (a *Menu) Update(ctx context.Context, recordID string, info map[string]interface{}) error {
	_, err := a.DB.UpdateByPK(a.TableName(),
		map[string]interface{}{"record_id": recordID},
		info)
	if err != nil {
		return errors.Wrap(err, "更新数据发生错误")
	}
	return nil
}

// Delete 删除数据
func (a *Menu) Delete(ctx context.Context, recordID string) error {
	_, err := a.DB.UpdateByPK(a.TableName(),
		map[string]interface{}{"record_id": recordID},
		map[string]interface{}{"deleted": time.Now().Unix()})
	if err != nil {
		return errors.Wrap(err, "删除数据发生错误")
	}
	return nil
}
