package cache

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	lru "github.com/hashicorp/golang-lru"
)

func TestLocalCodeCache_Set(t *testing.T) {
	testCase := []struct {
		name string
		mock func() *lru.Cache
		// 输入
		biz   string
		phone string
		code  string
		// 输出
		wantErr error
	}{
		{
			name: "验证码设置正确",
			mock: func() *lru.Cache {
				// 初始化一个新的本地缓存 ，达到每个测试用例单独，彼此不依赖的效果
				cache, err := lru.New(10)
				require.NoError(t, err)
				return cache
			},
			biz:     "login",
			phone:   "178xxxxxxx3",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "系统错误",
			mock: func() *lru.Cache {
				// 初始化一个新的本地缓存 ，达到每个测试用例单独，彼此不依赖的效果
				cache, err := lru.New(10)
				require.NoError(t, err)
				// 提前数据准备 ， “abc” 类型断言 codeItem 会失败
				cache.Add("phone_code:login:178xxxxxxx3", "abc")
				return cache
			},
			biz:     "login",
			phone:   "178xxxxxxx3",
			code:    "123456",
			wantErr: errors.New("系统错误"),
		},
		{
			name: "发送验证码太频繁",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", codeItem{
					code:   "123456",
					cnt:    3,
					expire: time.Now().Add(time.Minute*9 + time.Second*55),
				})
				return cache
			},
			biz:     "login",
			phone:   "178xxxxxxx3",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},

		{
			name: "重新设置验证码",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", codeItem{
					code:   "123456",
					cnt:    3,
					expire: time.Now().Add(time.Minute * 8),
				})
				return cache
			},
			biz:     "login",
			phone:   "178xxxxxxx3",
			code:    "123456",
			wantErr: nil,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			c := NewLocalCodeCache(tc.mock(), time.Minute*10)
			err := c.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}

}

func TestLocalCodeCache_Verify(t *testing.T) {
	testCase := []struct {
		name string
		mock func() *lru.Cache
		// 输入
		biz       string
		phone     string
		inputCode string
		// 输出
		wantBool bool
		wantErr  error
	}{
		{
			name: "验证码正确",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", codeItem{
					code:   "123456",
					cnt:    3,
					expire: time.Now().Add(time.Minute * 9),
				})
				return cache
			},
			biz:       "login",
			phone:     "178xxxxxxx3",
			inputCode: "123456",
			wantBool:  true,
			wantErr:   nil,
		},
		{
			name: "验证码缓存不存在",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				return cache
			},
			biz:       "login",
			phone:     "178xxxxxxx3",
			inputCode: "123456",
			wantBool:  false,
			wantErr:   ErrKeyNotExist,
		},
		{
			name: "系统错误",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", "abc")
				return cache
			},
			biz:       "login",
			phone:     "178xxxxxxx3",
			inputCode: "123456",
			wantBool:  false,
			wantErr:   errors.New("系统错误"),
		},
		{
			name: "验证码缓存过期",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", codeItem{
					code:   "123456",
					cnt:    3,
					expire: time.Now().Add(time.Minute * -1),
				})
				return cache
			},
			biz:       "login",
			phone:     "178xxxxxxx3",
			inputCode: "123456",
			wantBool:  false,
			wantErr:   ErrCodeVerifyTooManyTimes,
		},
		{
			name: "验证码次数太多",
			mock: func() *lru.Cache {
				cache, err := lru.New(10)
				require.NoError(t, err)
				cache.Add("phone_code:login:178xxxxxxx3", codeItem{
					code:   "123456",
					cnt:    0,
					expire: time.Now().Add(time.Minute * 9),
				})
				return cache
			},
			biz:       "login",
			phone:     "178xxxxxxx3",
			inputCode: "123456",
			wantBool:  false,
			wantErr:   ErrCodeVerifyTooManyTimes,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			c := NewLocalCodeCache(tc.mock(), time.Minute*10)
			ok, err := c.Verify(context.Background(), tc.biz, tc.phone, tc.inputCode)
			assert.Equal(t, tc.wantBool, ok)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
