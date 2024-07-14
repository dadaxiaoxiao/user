package service

import (
	"context"
	"errors"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/repository"
	repomocks "github.com/dadaxiaoxiao/user/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)


func Test_userService_Login(t *testing.T) {
	now := time.Now()
	testCase := []struct {
		name string
		mock func(controller *gomock.Controller) repository.UserRepository
		// 输入
		ctx      context.Context
		email    string
		password string
		// 输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "1426325504@qq.com").
					Return(domain.User{
						Id:       1,
						Email:    "1426325504@qq.com",
						Phone:    "178xxxxxxx3",
						Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
						Nickname: "yeqin",
						AboutMe:  "小胖子",
						Birthday: now,
						Ctime:    now,
					}, nil)
				return repo
			},
			ctx:      context.Background(),
			email:    "1426325504@qq.com",
			password: "hellword@123",
			wantUser: domain.User{
				Id:       1,
				Email:    "1426325504@qq.com",
				Phone:    "178xxxxxxx3",
				Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
				Nickname: "yeqin",
				AboutMe:  "小胖子",
				Birthday: now,
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户密码不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "1426325504@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			ctx:      context.Background(),
			email:    "1426325504@qq.com",
			password: "hellword@123",
			wantUser: domain.User{},
			// 返回密码错误
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "1426325504@qq.com").
					Return(domain.User{
						Id:       1,
						Email:    "1426325504@qq.com",
						Phone:    "178xxxxxxx3",
						Password: "$2a$10$mb97OEV00ZcyUl8ablHht.eJOKyMgOY/XcNLrBKzQGvTJDwJEb1Eq",
						Nickname: "yeqin",
						AboutMe:  "小胖子",
						Birthday: now,
						Ctime:    now,
					}, nil)
				return repo
			},
			ctx:   context.Background(),
			email: "1426325504@qq.com",
			// 输入一个错误的密码
			password: "hellword@12",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := NewUserService(tc.mock(ctrl), nil)
			user, err := service.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncrypted(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("hellword@123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(hash))
	}
}
