package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/repository/cache"
	cachemocks "github.com/dadaxiaoxiao/user/internal/repository/cache/mocks"
	"github.com/dadaxiaoxiao/user/internal/repository/dao"
	daomocks "github.com/dadaxiaoxiao/user/internal/repository/dao/mocks"
	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	// 去掉毫秒
	now = time.UnixMilli(now.UnixMilli())
	testCase := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache)

		//输入
		ctx context.Context
		id  int64

		// 输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(12)).
					Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDao(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(12)).Return(dao.User{
					Id: 12,
					Email: sql.NullString{
						String: "1426325504@qq.com",
						Valid:  true,
					},
					Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
					Nickname: sql.NullString{
						String: "yeqin",
						Valid:  true,
					},
					Phone: sql.NullString{
						String: "178xxxxxxx3",
						Valid:  true,
					},
					Birthday: sql.NullInt64{
						Int64: now.UnixMilli(),
						Valid: true,
					},
					AboutMe: sql.NullString{
						String: "一个灵活的小胖子",
						Valid:  true,
					},
					Ctime: now.UnixMilli(),
					Utime: now.UnixMilli(),
				}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       12,
					Email:    "1426325504@qq.com",
					Nickname: "yeqin",
					Phone:    "178xxxxxxx3",
					Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
					AboutMe:  "一个灵活的小胖子",
					Ctime:    now,
					Birthday: now,
				}).Return(nil)
				return d, c
			},

			ctx: context.Background(),
			id:  12,

			wantUser: domain.User{
				Id:       12,
				Email:    "1426325504@qq.com",
				Nickname: "yeqin",
				Phone:    "178xxxxxxx3",
				Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
				AboutMe:  "一个灵活的小胖子",
				Ctime:    now,
				Birthday: now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中，查询结果",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(12)).
					Return(domain.User{
						Id:       12,
						Email:    "1426325504@qq.com",
						Nickname: "yeqin",
						Phone:    "178xxxxxxx3",
						Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
						AboutMe:  "一个灵活的小胖子",
						Ctime:    now,
						Birthday: now,
					}, nil)
				d := daomocks.NewMockUserDao(ctrl)
				return d, c
			},

			ctx: context.Background(),
			id:  12,

			wantUser: domain.User{
				Id:       12,
				Email:    "1426325504@qq.com",
				Nickname: "yeqin",
				Phone:    "178xxxxxxx3",
				Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
				AboutMe:  "一个灵活的小胖子",
				Ctime:    now,
				Birthday: now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，查询数据库异常",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(12)).
					Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDao(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(12)).Return(dao.User{}, errors.New("db 异常"))

				return d, c
			},
			ctx:      context.Background(),
			id:       12,
			wantUser: domain.User{},
			wantErr:  errors.New("db 异常"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := NewCachedUserRepository(tc.mock(ctrl))
			user, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
