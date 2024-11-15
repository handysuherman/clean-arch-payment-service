// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/token/maker.go
//
// Generated by this command:
//
//	mockgen -package tokenMock -destination pkg/token/mock/mock.go -source=pkg/token/maker.go
//

// Package tokenMock is a generated GoMock package.
package tokenMock

import (
	reflect "reflect"
	time "time"

	token "github.com/handysuherman/clean-arch-payment-service/internal/pkg/token"
	gomock "go.uber.org/mock/gomock"
)

// MockMaker is a mock of Maker interface.
type MockMaker struct {
	ctrl     *gomock.Controller
	recorder *MockMakerMockRecorder
}

// MockMakerMockRecorder is the mock recorder for MockMaker.
type MockMakerMockRecorder struct {
	mock *MockMaker
}

// NewMockMaker creates a new mock instance.
func NewMockMaker(ctrl *gomock.Controller) *MockMaker {
	mock := &MockMaker{ctrl: ctrl}
	mock.recorder = &MockMakerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMaker) EXPECT() *MockMakerMockRecorder {
	return m.recorder
}

// CreateToken mocks base method.
func (m *MockMaker) CreateToken(claimer *token.Claimer, duration time.Duration, tokenType string) (string, *token.Payload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateToken", claimer, duration, tokenType)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*token.Payload)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateToken indicates an expected call of CreateToken.
func (mr *MockMakerMockRecorder) CreateToken(claimer, duration, tokenType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateToken", reflect.TypeOf((*MockMaker)(nil).CreateToken), claimer, duration, tokenType)
}

// VerifyToken mocks base method.
func (m *MockMaker) VerifyToken(tokens, expectedTokenType string) (*token.Payload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyToken", tokens, expectedTokenType)
	ret0, _ := ret[0].(*token.Payload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyToken indicates an expected call of VerifyToken.
func (mr *MockMakerMockRecorder) VerifyToken(tokens, expectedTokenType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyToken", reflect.TypeOf((*MockMaker)(nil).VerifyToken), tokens, expectedTokenType)
}
