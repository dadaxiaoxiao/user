package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCase := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB
		// 输入
		ctx  context.Context
		user User
		// 输出
		wantErr error
	}{
		{
			name: "插入成功",
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "1426325504@qq.com",
					Valid:  true,
				},
			},
		},
		{
			name: "邮箱冲突",
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				require.NoError(t, err)
				return mockDB
			},
			user:    User{},
			wantErr: ErrUserDuplicateEmail,
		},
		{
			name: "数据库错误",
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("数据库错误"))
				require.NoError(t, err)
				return mockDB
			},
			user:    User{},
			wantErr: errors.New("数据库错误"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.sqlmock(t)
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: sqlDB,
				// SELECT VERSION;
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			// 初始化db不能出错 ，断言必须为nil
			assert.NoError(t, err)
			dao := NewGORMUserDAO(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
