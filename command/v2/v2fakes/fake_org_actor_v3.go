// This file was generated by counterfeiter
package v2fakes

import (
	"sync"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/v2"
)

type FakeOrgActorV3 struct {
	GetIsolationSegmentsByOrganizationStub        func(orgName string) ([]v3action.IsolationSegment, v3action.Warnings, error)
	getIsolationSegmentsByOrganizationMutex       sync.RWMutex
	getIsolationSegmentsByOrganizationArgsForCall []struct {
		orgName string
	}
	getIsolationSegmentsByOrganizationReturns struct {
		result1 []v3action.IsolationSegment
		result2 v3action.Warnings
		result3 error
	}
	getIsolationSegmentsByOrganizationReturnsOnCall map[int]struct {
		result1 []v3action.IsolationSegment
		result2 v3action.Warnings
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOrgActorV3) GetIsolationSegmentsByOrganization(orgName string) ([]v3action.IsolationSegment, v3action.Warnings, error) {
	fake.getIsolationSegmentsByOrganizationMutex.Lock()
	ret, specificReturn := fake.getIsolationSegmentsByOrganizationReturnsOnCall[len(fake.getIsolationSegmentsByOrganizationArgsForCall)]
	fake.getIsolationSegmentsByOrganizationArgsForCall = append(fake.getIsolationSegmentsByOrganizationArgsForCall, struct {
		orgName string
	}{orgName})
	fake.recordInvocation("GetIsolationSegmentsByOrganization", []interface{}{orgName})
	fake.getIsolationSegmentsByOrganizationMutex.Unlock()
	if fake.GetIsolationSegmentsByOrganizationStub != nil {
		return fake.GetIsolationSegmentsByOrganizationStub(orgName)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fake.getIsolationSegmentsByOrganizationReturns.result1, fake.getIsolationSegmentsByOrganizationReturns.result2, fake.getIsolationSegmentsByOrganizationReturns.result3
}

func (fake *FakeOrgActorV3) GetIsolationSegmentsByOrganizationCallCount() int {
	fake.getIsolationSegmentsByOrganizationMutex.RLock()
	defer fake.getIsolationSegmentsByOrganizationMutex.RUnlock()
	return len(fake.getIsolationSegmentsByOrganizationArgsForCall)
}

func (fake *FakeOrgActorV3) GetIsolationSegmentsByOrganizationArgsForCall(i int) string {
	fake.getIsolationSegmentsByOrganizationMutex.RLock()
	defer fake.getIsolationSegmentsByOrganizationMutex.RUnlock()
	return fake.getIsolationSegmentsByOrganizationArgsForCall[i].orgName
}

func (fake *FakeOrgActorV3) GetIsolationSegmentsByOrganizationReturns(result1 []v3action.IsolationSegment, result2 v3action.Warnings, result3 error) {
	fake.GetIsolationSegmentsByOrganizationStub = nil
	fake.getIsolationSegmentsByOrganizationReturns = struct {
		result1 []v3action.IsolationSegment
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeOrgActorV3) GetIsolationSegmentsByOrganizationReturnsOnCall(i int, result1 []v3action.IsolationSegment, result2 v3action.Warnings, result3 error) {
	fake.GetIsolationSegmentsByOrganizationStub = nil
	if fake.getIsolationSegmentsByOrganizationReturnsOnCall == nil {
		fake.getIsolationSegmentsByOrganizationReturnsOnCall = make(map[int]struct {
			result1 []v3action.IsolationSegment
			result2 v3action.Warnings
			result3 error
		})
	}
	fake.getIsolationSegmentsByOrganizationReturnsOnCall[i] = struct {
		result1 []v3action.IsolationSegment
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeOrgActorV3) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getIsolationSegmentsByOrganizationMutex.RLock()
	defer fake.getIsolationSegmentsByOrganizationMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeOrgActorV3) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ v2.OrgActorV3 = new(FakeOrgActorV3)
