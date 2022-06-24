// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package machinemanager_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	basetesting "github.com/juju/juju/api/base/testing"
	"github.com/juju/juju/api/client/machinemanager"
	"github.com/juju/juju/rpc/params"
	"github.com/juju/juju/storage"
	coretesting "github.com/juju/juju/testing"
)

var _ = gc.Suite(&MachinemanagerSuite{})

type MachinemanagerSuite struct {
	coretesting.BaseSuite
}

func newClient(f basetesting.APICallerFunc) *machinemanager.Client {
	return machinemanager.NewClient(f)
}

func (s *MachinemanagerSuite) TestAddMachines(c *gc.C) {
	apiResult := []params.AddMachinesResult{
		{Machine: "machine-1", Error: nil},
		{Machine: "machine-2", Error: nil},
	}

	var callCount int
	st := newClient(func(objType string, version int, id, request string, arg, result interface{}) error {
		c.Check(objType, gc.Equals, "MachineManager")
		c.Check(version, gc.Equals, 0)
		c.Check(id, gc.Equals, "")
		c.Check(request, gc.Equals, "AddMachines")
		c.Check(arg, gc.DeepEquals, params.AddMachines{
			MachineParams: []params.AddMachineParams{
				{
					Series: "trusty",
					Disks:  []storage.Constraints{{Pool: "loop", Size: 1}},
				},
				{
					Series: "precise",
				},
			},
		})
		c.Assert(result, gc.FitsTypeOf, &params.AddMachinesResults{})
		*(result.(*params.AddMachinesResults)) = params.AddMachinesResults{
			Machines: apiResult,
		}
		callCount++
		return nil
	})

	machines := []params.AddMachineParams{{
		Series: "trusty",
		Disks:  []storage.Constraints{{Pool: "loop", Size: 1}},
	}, {
		Series: "precise",
	}}
	result, err := st.AddMachines(machines)
	c.Check(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, apiResult)
	c.Check(callCount, gc.Equals, 1)
}

func (s *MachinemanagerSuite) TestAddMachinesClientError(c *gc.C) {
	st := newClient(func(objType string, version int, id, request string, arg, result interface{}) error {
		return errors.New("blargh")
	})
	_, err := st.AddMachines([]params.AddMachineParams{{Series: "focal"}})
	c.Check(err, gc.ErrorMatches, "blargh")
}

func (s *MachinemanagerSuite) TestAddMachinesServerError(c *gc.C) {
	apiResult := []params.AddMachinesResult{{
		Error: &params.Error{Message: "MSG", Code: "621"},
	}}

	st := newClient(func(objType string, version int, id, request string, arg, result interface{}) error {
		*(result.(*params.AddMachinesResults)) = params.AddMachinesResults{
			Machines: apiResult,
		}
		return nil
	})
	machines := []params.AddMachineParams{{
		Series: "trusty",
	}}
	results, err := st.AddMachines(machines)
	c.Check(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, apiResult)
}

func (s *MachinemanagerSuite) TestAddMachinesResultCountInvalid(c *gc.C) {
	for _, n := range []int{0, 2} {
		st := newClient(func(objType string, version int, id, request string, arg, result interface{}) error {
			var results []params.AddMachinesResult
			for i := 0; i < n; i++ {
				results = append(results, params.AddMachinesResult{
					Error: &params.Error{Message: "MSG", Code: "621"},
				})
			}
			*(result.(*params.AddMachinesResults)) = params.AddMachinesResults{Machines: results}
			return nil
		})
		machines := []params.AddMachineParams{{
			Series: "trusty",
		}}
		_, err := st.AddMachines(machines)
		c.Check(err, gc.ErrorMatches, fmt.Sprintf("expected 1 result, got %d", n))
	}
}

func (s *MachinemanagerSuite) TestRetryProvisioning(c *gc.C) {
	client := machinemanager.NewClient(
		basetesting.BestVersionCaller{
			BestVersion: 7,
			APICallerFunc: basetesting.APICallerFunc(func(objType string, version int, id, request string, a, response interface{}) error {
				c.Assert(request, gc.Equals, "RetryProvisioning")
				c.Assert(version, gc.Equals, 7)
				c.Assert(a, jc.DeepEquals, params.Entities{
					Entities: []params.Entity{
						{Tag: "machine-0"},
						{Tag: "machine-1"},
					},
				})
				c.Assert(response, gc.FitsTypeOf, &params.ErrorResults{})
				out := response.(*params.ErrorResults)
				*out = params.ErrorResults{Results: []params.ErrorResult{
					{Error: &params.Error{Code: "boom"}},
					{}},
				}
				return nil
			})})
	result, err := client.RetryProvisioning(names.NewMachineTag("0"), names.NewMachineTag("1"))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, []params.ErrorResult{
		{&params.Error{Code: "boom"}},
		{},
	})
}

func (s *MachinemanagerSuite) TestProvisioningScript(c *gc.C) {
	client := machinemanager.NewClient(
		basetesting.BestVersionCaller{
			BestVersion: 7,
			APICallerFunc: basetesting.APICallerFunc(func(objType string, version int, id, request string, a, response interface{}) error {
				c.Assert(request, gc.Equals, "ProvisioningScript")
				c.Assert(version, gc.Equals, 7)
				c.Assert(a, jc.DeepEquals, params.ProvisioningScriptParams{
					MachineId:              "0",
					Nonce:                  "nonce",
					DataDir:                "/path/to/data",
					DisablePackageCommands: true,
				})
				c.Assert(response, gc.FitsTypeOf, &params.ProvisioningScriptResult{})
				out := response.(*params.ProvisioningScriptResult)
				*out = params.ProvisioningScriptResult{Script: "script"}
				return nil
			})})
	script, err := client.ProvisioningScript(params.ProvisioningScriptParams{
		MachineId:              "0",
		Nonce:                  "nonce",
		DataDir:                "/path/to/data",
		DisablePackageCommands: true,
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(script, gc.Equals, "script")
}

func (s *MachinemanagerSuite) clientToTestDestroyMachinesWithParams(c *gc.C, v int, maxWait *time.Duration) (*machinemanager.Client, []params.DestroyMachineResult) {
	expectedResults := []params.DestroyMachineResult{{
		Error: &params.Error{Message: "boo"},
	}, {
		Info: &params.DestroyMachineInfo{
			DestroyedUnits:   []params.Entity{{Tag: "unit-foo-0"}},
			DestroyedStorage: []params.Entity{{Tag: "storage-pgdata-0"}},
			DetachedStorage:  []params.Entity{{Tag: "storage-pgdata-1"}},
		},
	}}
	client := machinemanager.NewClient(
		basetesting.BestVersionCaller{
			BestVersion: v,
			APICallerFunc: basetesting.APICallerFunc(func(objType string, version int, id, request string, a, response interface{}) error {
				c.Assert(request, gc.Equals, "DestroyMachineWithParams")
				c.Assert(version, gc.Equals, v)
				c.Assert(a, jc.DeepEquals, params.DestroyMachinesParams{
					Keep:  true,
					Force: true,
					MachineTags: []string{
						"machine-0",
						"machine-0-lxd-1",
					},
					MaxWait: maxWait,
				})
				c.Assert(response, gc.FitsTypeOf, &params.DestroyMachineResults{})
				out := response.(*params.DestroyMachineResults)
				*out = params.DestroyMachineResults{Results: expectedResults}
				return nil
			})})
	return client, expectedResults
}

func (s *MachinemanagerSuite) TestDestroyMachinesWithParamsV5NoWait(c *gc.C) {
	// MaxWait will be ignored in all versions < 6, so expect the argument
	// to apiserver to always be nl.
	client, expected := s.clientToTestDestroyMachinesWithParams(c, 5, (*time.Duration)(nil))
	noWait := 0 * time.Second
	results, err := client.DestroyMachinesWithParams(true, true, &noWait, "0", "0/lxd/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, expected)
}

func (s *MachinemanagerSuite) TestDestroyMachinesWithParamsV5NilWait(c *gc.C) {
	// MaxWait will be ignored in all versions < 6, so expect the argument
	// to apiserver to always be nl.
	client, expected := s.clientToTestDestroyMachinesWithParams(c, 5, (*time.Duration)(nil))
	results, err := client.DestroyMachinesWithParams(true, true, (*time.Duration)(nil), "0", "0/lxd/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, expected)
}

func (s *MachinemanagerSuite) TestDestroyMachinesWithParamsNoWait(c *gc.C) {
	noWait := 0 * time.Second
	client, expected := s.clientToTestDestroyMachinesWithParams(c, 6, &noWait)
	results, err := client.DestroyMachinesWithParams(true, true, &noWait, "0", "0/lxd/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, expected)
}

func (s *MachinemanagerSuite) TestDestroyMachinesWithParamsNilWait(c *gc.C) {
	client, expected := s.clientToTestDestroyMachinesWithParams(c, 6, (*time.Duration)(nil))
	results, err := client.DestroyMachinesWithParams(true, true, (*time.Duration)(nil), "0", "0/lxd/1")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, expected)
}
