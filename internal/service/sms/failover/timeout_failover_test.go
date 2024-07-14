package failover

import (
	"context"
	"errors"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	smsmocks "github.com/dadaxiaoxiao/user/internal/service/sms/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestTimeoutFailOverSMSService_Send(t *testing.T) {
	testCase := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []sms.Service
		threshold int32
		// 通过控制私有变量，模拟各种场景
		idx int32
		cnt int32

		wantErr error
		wantIdx int32
		wantCnt int32
	}{
		{
			name: "触发阈值，切换后成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			idx:       0,
			cnt:       3,
			wantErr:   nil,
			// 切换到1
			wantIdx: 1,
			// 重置计算
			wantCnt: 0,
		},
		{
			name: "触发阈值，切换后依旧超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			idx:       0,
			cnt:       3,
			wantErr:   context.DeadlineExceeded,
			// 切换到1
			wantIdx: 1,
			// 重置计算
			wantCnt: 1,
		},
		{
			name: "触发阈值，切换后失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("系统错误"))
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			idx:       0,
			cnt:       3,
			wantErr:   errors.New("系统错误"),
			// 切换到1
			wantIdx: 1,
			// 重置计算
			wantCnt: 0,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewTimeoutFailoverSMSService(tc.mock(ctrl), tc.threshold)
			svc.idx = tc.idx
			svc.cnt = tc.cnt

			err := svc.Send(context.Background(), "1932694", []string{"123456"}, "17875513413")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantIdx, svc.idx)
			assert.Equal(t, tc.wantCnt, svc.cnt)

		})
	}
}
