// Code generated by MockGen. DO NOT EDIT.
// Source: ../eks.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	eks "github.com/aws/aws-sdk-go/service/eks"
	gomock "github.com/golang/mock/gomock"
)

// MockEKSServiceInterface is a mock of EKSServiceInterface interface.
type MockEKSServiceInterface struct {
	ctrl     *gomock.Controller
	recorder *MockEKSServiceInterfaceMockRecorder
}

// MockEKSServiceInterfaceMockRecorder is the mock recorder for MockEKSServiceInterface.
type MockEKSServiceInterfaceMockRecorder struct {
	mock *MockEKSServiceInterface
}

// NewMockEKSServiceInterface creates a new mock instance.
func NewMockEKSServiceInterface(ctrl *gomock.Controller) *MockEKSServiceInterface {
	mock := &MockEKSServiceInterface{ctrl: ctrl}
	mock.recorder = &MockEKSServiceInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEKSServiceInterface) EXPECT() *MockEKSServiceInterfaceMockRecorder {
	return m.recorder
}

// CreateAddon mocks base method.
func (m *MockEKSServiceInterface) CreateAddon(input *eks.CreateAddonInput) (*eks.CreateAddonOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAddon", input)
	ret0, _ := ret[0].(*eks.CreateAddonOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAddon indicates an expected call of CreateAddon.
func (mr *MockEKSServiceInterfaceMockRecorder) CreateAddon(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAddon", reflect.TypeOf((*MockEKSServiceInterface)(nil).CreateAddon), input)
}

// CreateCluster mocks base method.
func (m *MockEKSServiceInterface) CreateCluster(input *eks.CreateClusterInput) (*eks.CreateClusterOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCluster", input)
	ret0, _ := ret[0].(*eks.CreateClusterOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCluster indicates an expected call of CreateCluster.
func (mr *MockEKSServiceInterfaceMockRecorder) CreateCluster(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCluster", reflect.TypeOf((*MockEKSServiceInterface)(nil).CreateCluster), input)
}

// CreateNodegroup mocks base method.
func (m *MockEKSServiceInterface) CreateNodegroup(input *eks.CreateNodegroupInput) (*eks.CreateNodegroupOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNodegroup", input)
	ret0, _ := ret[0].(*eks.CreateNodegroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateNodegroup indicates an expected call of CreateNodegroup.
func (mr *MockEKSServiceInterfaceMockRecorder) CreateNodegroup(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNodegroup", reflect.TypeOf((*MockEKSServiceInterface)(nil).CreateNodegroup), input)
}

// DeleteCluster mocks base method.
func (m *MockEKSServiceInterface) DeleteCluster(input *eks.DeleteClusterInput) (*eks.DeleteClusterOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCluster", input)
	ret0, _ := ret[0].(*eks.DeleteClusterOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteCluster indicates an expected call of DeleteCluster.
func (mr *MockEKSServiceInterfaceMockRecorder) DeleteCluster(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCluster", reflect.TypeOf((*MockEKSServiceInterface)(nil).DeleteCluster), input)
}

// DeleteNodegroup mocks base method.
func (m *MockEKSServiceInterface) DeleteNodegroup(input *eks.DeleteNodegroupInput) (*eks.DeleteNodegroupOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNodegroup", input)
	ret0, _ := ret[0].(*eks.DeleteNodegroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteNodegroup indicates an expected call of DeleteNodegroup.
func (mr *MockEKSServiceInterfaceMockRecorder) DeleteNodegroup(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNodegroup", reflect.TypeOf((*MockEKSServiceInterface)(nil).DeleteNodegroup), input)
}

// DescribeAddon mocks base method.
func (m *MockEKSServiceInterface) DescribeAddon(input *eks.DescribeAddonInput) (*eks.DescribeAddonOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeAddon", input)
	ret0, _ := ret[0].(*eks.DescribeAddonOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeAddon indicates an expected call of DescribeAddon.
func (mr *MockEKSServiceInterfaceMockRecorder) DescribeAddon(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeAddon", reflect.TypeOf((*MockEKSServiceInterface)(nil).DescribeAddon), input)
}

// DescribeCluster mocks base method.
func (m *MockEKSServiceInterface) DescribeCluster(input *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeCluster", input)
	ret0, _ := ret[0].(*eks.DescribeClusterOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeCluster indicates an expected call of DescribeCluster.
func (mr *MockEKSServiceInterfaceMockRecorder) DescribeCluster(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeCluster", reflect.TypeOf((*MockEKSServiceInterface)(nil).DescribeCluster), input)
}

// DescribeNodegroup mocks base method.
func (m *MockEKSServiceInterface) DescribeNodegroup(input *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeNodegroup", input)
	ret0, _ := ret[0].(*eks.DescribeNodegroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeNodegroup indicates an expected call of DescribeNodegroup.
func (mr *MockEKSServiceInterfaceMockRecorder) DescribeNodegroup(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeNodegroup", reflect.TypeOf((*MockEKSServiceInterface)(nil).DescribeNodegroup), input)
}

// ListClusters mocks base method.
func (m *MockEKSServiceInterface) ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListClusters", input)
	ret0, _ := ret[0].(*eks.ListClustersOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListClusters indicates an expected call of ListClusters.
func (mr *MockEKSServiceInterfaceMockRecorder) ListClusters(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListClusters", reflect.TypeOf((*MockEKSServiceInterface)(nil).ListClusters), input)
}

// ListNodegroups mocks base method.
func (m *MockEKSServiceInterface) ListNodegroups(input *eks.ListNodegroupsInput) (*eks.ListNodegroupsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListNodegroups", input)
	ret0, _ := ret[0].(*eks.ListNodegroupsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListNodegroups indicates an expected call of ListNodegroups.
func (mr *MockEKSServiceInterfaceMockRecorder) ListNodegroups(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListNodegroups", reflect.TypeOf((*MockEKSServiceInterface)(nil).ListNodegroups), input)
}

// TagResource mocks base method.
func (m *MockEKSServiceInterface) TagResource(input *eks.TagResourceInput) (*eks.TagResourceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TagResource", input)
	ret0, _ := ret[0].(*eks.TagResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TagResource indicates an expected call of TagResource.
func (mr *MockEKSServiceInterfaceMockRecorder) TagResource(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TagResource", reflect.TypeOf((*MockEKSServiceInterface)(nil).TagResource), input)
}

// UntagResource mocks base method.
func (m *MockEKSServiceInterface) UntagResource(input *eks.UntagResourceInput) (*eks.UntagResourceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UntagResource", input)
	ret0, _ := ret[0].(*eks.UntagResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UntagResource indicates an expected call of UntagResource.
func (mr *MockEKSServiceInterfaceMockRecorder) UntagResource(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UntagResource", reflect.TypeOf((*MockEKSServiceInterface)(nil).UntagResource), input)
}

// UpdateClusterConfig mocks base method.
func (m *MockEKSServiceInterface) UpdateClusterConfig(input *eks.UpdateClusterConfigInput) (*eks.UpdateClusterConfigOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateClusterConfig", input)
	ret0, _ := ret[0].(*eks.UpdateClusterConfigOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateClusterConfig indicates an expected call of UpdateClusterConfig.
func (mr *MockEKSServiceInterfaceMockRecorder) UpdateClusterConfig(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateClusterConfig", reflect.TypeOf((*MockEKSServiceInterface)(nil).UpdateClusterConfig), input)
}

// UpdateClusterVersion mocks base method.
func (m *MockEKSServiceInterface) UpdateClusterVersion(input *eks.UpdateClusterVersionInput) (*eks.UpdateClusterVersionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateClusterVersion", input)
	ret0, _ := ret[0].(*eks.UpdateClusterVersionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateClusterVersion indicates an expected call of UpdateClusterVersion.
func (mr *MockEKSServiceInterfaceMockRecorder) UpdateClusterVersion(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateClusterVersion", reflect.TypeOf((*MockEKSServiceInterface)(nil).UpdateClusterVersion), input)
}

// UpdateNodegroupConfig mocks base method.
func (m *MockEKSServiceInterface) UpdateNodegroupConfig(input *eks.UpdateNodegroupConfigInput) (*eks.UpdateNodegroupConfigOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNodegroupConfig", input)
	ret0, _ := ret[0].(*eks.UpdateNodegroupConfigOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNodegroupConfig indicates an expected call of UpdateNodegroupConfig.
func (mr *MockEKSServiceInterfaceMockRecorder) UpdateNodegroupConfig(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNodegroupConfig", reflect.TypeOf((*MockEKSServiceInterface)(nil).UpdateNodegroupConfig), input)
}

// UpdateNodegroupVersion mocks base method.
func (m *MockEKSServiceInterface) UpdateNodegroupVersion(input *eks.UpdateNodegroupVersionInput) (*eks.UpdateNodegroupVersionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNodegroupVersion", input)
	ret0, _ := ret[0].(*eks.UpdateNodegroupVersionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNodegroupVersion indicates an expected call of UpdateNodegroupVersion.
func (mr *MockEKSServiceInterfaceMockRecorder) UpdateNodegroupVersion(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNodegroupVersion", reflect.TypeOf((*MockEKSServiceInterface)(nil).UpdateNodegroupVersion), input)
}
