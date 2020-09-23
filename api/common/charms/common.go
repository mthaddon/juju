// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

// charms provides a client for accessing the charms API.
package charms

import (
	"github.com/juju/charm/v8"
	"github.com/juju/charm/v8/resource"
	"github.com/juju/errors"
	"github.com/juju/version"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/apiserver/params"
)

// CharmsClient allows access to the charms API end point.
type CharmsClient struct {
	facade base.FacadeCaller
}

// NewCharmsClient creates a new client for accessing the charms API.
func NewCharmsClient(facade base.FacadeCaller) *CharmsClient {
	return &CharmsClient{facade: facade}
}

// CharmInfo holds information about a charm.
type CharmInfo struct {
	Revision   int
	URL        string
	Config     *charm.Config
	Meta       *charm.Meta
	Actions    *charm.Actions
	Metrics    *charm.Metrics
	LXDProfile *charm.LXDProfile
}

func (info *CharmInfo) Charm() charm.Charm {
	return &charmImpl{info}
}

// CharmInfo returns information about the requested charm.
func (c *CharmsClient) CharmInfo(charmURL string) (*CharmInfo, error) {
	args := params.CharmURL{URL: charmURL}
	var info params.Charm
	if err := c.facade.FacadeCall("CharmInfo", args, &info); err != nil {
		return nil, errors.Trace(err)
	}
	meta, err := convertCharmMeta(info.Meta)
	if err != nil {
		return nil, errors.Trace(err)
	}
	result := &CharmInfo{
		Revision:   info.Revision,
		URL:        info.URL,
		Config:     params.FromCharmOptionMap(info.Config),
		Meta:       meta,
		Actions:    convertCharmActions(info.Actions),
		Metrics:    convertCharmMetrics(info.Metrics),
		LXDProfile: convertCharmLXDProfile(info.LXDProfile),
	}
	return result, nil
}

func convertCharmMeta(meta *params.CharmMeta) (*charm.Meta, error) {
	if meta == nil {
		return nil, nil
	}
	minVersion, err := version.Parse(meta.MinJujuVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}
	resources, err := convertCharmResourceMetaMap(meta.Resources)
	if err != nil {
		return nil, errors.Trace(err)
	}
	result := &charm.Meta{
		Name:           meta.Name,
		Summary:        meta.Summary,
		Description:    meta.Description,
		Subordinate:    meta.Subordinate,
		Provides:       convertCharmRelationMap(meta.Provides),
		Requires:       convertCharmRelationMap(meta.Requires),
		Peers:          convertCharmRelationMap(meta.Peers),
		ExtraBindings:  convertCharmExtraBindingMap(meta.ExtraBindings),
		Categories:     meta.Categories,
		Tags:           meta.Tags,
		Series:         meta.Series,
		Storage:        convertCharmStorageMap(meta.Storage),
		Deployment:     convertCharmDeployment(meta.Deployment),
		Devices:        convertCharmDevices(meta.Devices),
		PayloadClasses: convertCharmPayloadClassMap(meta.PayloadClasses),
		Resources:      resources,
		Terms:          meta.Terms,
		MinJujuVersion: minVersion,
		Systems:        convertCharmSystems(meta.Systems),
		Platforms:      convertCharmPlatforms(meta.Platforms),
		Architectures:  convertCharmArchitectures(meta.Architectures),
		Containers:     convertCharmContainers(meta.Containers),
	}
	return result, nil
}

func convertCharmRelationMap(relations map[string]params.CharmRelation) map[string]charm.Relation {
	if len(relations) == 0 {
		return nil
	}
	result := make(map[string]charm.Relation)
	for key, value := range relations {
		result[key] = convertCharmRelation(value)
	}
	return result
}

func convertCharmRelation(relation params.CharmRelation) charm.Relation {
	return charm.Relation{
		Name:      relation.Name,
		Role:      charm.RelationRole(relation.Role),
		Interface: relation.Interface,
		Optional:  relation.Optional,
		Limit:     relation.Limit,
		Scope:     charm.RelationScope(relation.Scope),
	}
}

func convertCharmStorageMap(storage map[string]params.CharmStorage) map[string]charm.Storage {
	if len(storage) == 0 {
		return nil
	}
	result := make(map[string]charm.Storage)
	for key, value := range storage {
		result[key] = convertCharmStorage(value)
	}
	return result
}

func convertCharmStorage(storage params.CharmStorage) charm.Storage {
	return charm.Storage{
		Name:        storage.Name,
		Description: storage.Description,
		Type:        charm.StorageType(storage.Type),
		Shared:      storage.Shared,
		ReadOnly:    storage.ReadOnly,
		CountMin:    storage.CountMin,
		CountMax:    storage.CountMax,
		MinimumSize: storage.MinimumSize,
		Location:    storage.Location,
		Properties:  storage.Properties,
	}
}

func convertCharmPayloadClassMap(payload map[string]params.CharmPayloadClass) map[string]charm.PayloadClass {
	if len(payload) == 0 {
		return nil
	}
	result := make(map[string]charm.PayloadClass)
	for key, value := range payload {
		result[key] = convertCharmPayloadClass(value)
	}
	return result
}

func convertCharmPayloadClass(payload params.CharmPayloadClass) charm.PayloadClass {
	return charm.PayloadClass{
		Name: payload.Name,
		Type: payload.Type,
	}
}

func convertCharmResourceMetaMap(resources map[string]params.CharmResourceMeta) (map[string]resource.Meta, error) {
	if len(resources) == 0 {
		return nil, nil
	}
	result := make(map[string]resource.Meta)
	for key, value := range resources {
		converted, err := convertCharmResourceMeta(value)
		if err != nil {
			return nil, errors.Trace(err)
		}
		result[key] = converted
	}
	return result, nil
}

func convertCharmResourceMeta(meta params.CharmResourceMeta) (resource.Meta, error) {
	resourceType, err := resource.ParseType(meta.Type)
	if err != nil {
		return resource.Meta{}, errors.Trace(err)
	}
	return resource.Meta{
		Name:        meta.Name,
		Type:        resourceType,
		Path:        meta.Path,
		Description: meta.Description,
	}, nil
}

func convertCharmActions(actions *params.CharmActions) *charm.Actions {
	if actions == nil {
		return nil
	}
	return &charm.Actions{
		ActionSpecs: convertCharmActionSpecMap(actions.ActionSpecs),
	}
}

func convertCharmActionSpecMap(specs map[string]params.CharmActionSpec) map[string]charm.ActionSpec {
	if len(specs) == 0 {
		return nil
	}
	result := make(map[string]charm.ActionSpec)
	for key, value := range specs {
		result[key] = convertCharmActionSpec(value)
	}
	return result
}

func convertCharmActionSpec(spec params.CharmActionSpec) charm.ActionSpec {
	return charm.ActionSpec{
		Description: spec.Description,
		Params:      spec.Params,
	}
}

func convertCharmMetrics(metrics *params.CharmMetrics) *charm.Metrics {
	if metrics == nil {
		return nil
	}
	return &charm.Metrics{
		Metrics: convertCharmMetricMap(metrics.Metrics),
		Plan:    convertCharmPlan(metrics.Plan),
	}
}

func convertCharmPlan(plan params.CharmPlan) *charm.Plan {
	return &charm.Plan{Required: plan.Required}
}

func convertCharmMetricMap(metrics map[string]params.CharmMetric) map[string]charm.Metric {
	if len(metrics) == 0 {
		return nil
	}
	result := make(map[string]charm.Metric)
	for key, value := range metrics {
		result[key] = convertCharmMetric(value)
	}
	return result
}

func convertCharmMetric(metric params.CharmMetric) charm.Metric {
	return charm.Metric{
		Type:        charm.MetricType(metric.Type),
		Description: metric.Description,
	}
}

func convertCharmExtraBindingMap(bindings map[string]string) map[string]charm.ExtraBinding {
	if len(bindings) == 0 {
		return nil
	}
	result := make(map[string]charm.ExtraBinding)
	for key, value := range bindings {
		result[key] = charm.ExtraBinding{value}
	}
	return result
}

func convertCharmLXDProfile(lxdProfile *params.CharmLXDProfile) *charm.LXDProfile {
	if lxdProfile == nil {
		return nil
	}
	return &charm.LXDProfile{
		Description: lxdProfile.Description,
		Config:      convertCharmLXDProfileConfigMap(lxdProfile.Config),
		Devices:     convertCharmLXDProfileDevicesMap(lxdProfile.Devices),
	}
}

func convertCharmLXDProfileConfigMap(config map[string]string) map[string]string {
	result := make(map[string]string, len(config))
	for k, v := range config {
		result[k] = v
	}
	return result
}

func convertCharmLXDProfileDevicesMap(devices map[string]map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(devices))
	for k, v := range devices {
		nested := make(map[string]string, len(v))
		for nk, nv := range v {
			nested[nk] = nv
		}
		result[k] = nested
	}
	return result
}

func convertCharmDeployment(deployment *params.CharmDeployment) *charm.Deployment {
	if deployment == nil {
		return nil
	}
	return &charm.Deployment{
		DeploymentType: charm.DeploymentType(deployment.DeploymentType),
		DeploymentMode: charm.DeploymentMode(deployment.DeploymentMode),
		ServiceType:    charm.ServiceType(deployment.ServiceType),
		MinVersion:     deployment.MinVersion,
	}
}

func convertCharmDevices(devices map[string]params.CharmDevice) map[string]charm.Device {
	if devices == nil {
		return nil
	}
	results := make(map[string]charm.Device)
	for k, v := range devices {
		results[k] = charm.Device{
			Name:        v.Name,
			Description: v.Description,
			Type:        charm.DeviceType(v.Type),
			CountMin:    v.CountMin,
			CountMax:    v.CountMax,
		}
	}
	return results
}

type charmImpl struct {
	info *CharmInfo
}

func (c *charmImpl) Meta() *charm.Meta {
	return c.info.Meta
}

func (c *charmImpl) Config() *charm.Config {
	return c.info.Config
}

func (c *charmImpl) Metrics() *charm.Metrics {
	return c.info.Metrics
}

func (c *charmImpl) Actions() *charm.Actions {
	return c.info.Actions
}

func (c *charmImpl) Revision() int {
	return c.info.Revision
}

func convertCharmSystems(input []params.CharmSystem) []charm.System {
	systems := []charm.System(nil)
	for _, v := range input {
		systems = append(systems, charm.System{
			OS:       v.OS,
			Version:  v.Version,
			Resource: v.Resource,
		})
	}
	return systems
}

func convertCharmPlatforms(input []string) []charm.Platform {
	platforms := []charm.Platform(nil)
	for _, v := range input {
		platforms = append(platforms, charm.Platform(v))
	}
	return platforms
}

func convertCharmArchitectures(input []string) []charm.Architecture {
	architectures := []charm.Architecture(nil)
	for _, v := range input {
		architectures = append(architectures, charm.Architecture(v))
	}
	return architectures
}

func convertCharmContainers(input map[string]params.CharmContainer) map[string]charm.Container {
	containers := map[string]charm.Container{}
	for k, v := range input {
		containers[k] = charm.Container{
			Systems: convertCharmSystems(v.Systems),
			Mounts:  convertCharmMounts(v.Mounts),
		}
	}
	if len(containers) == 0 {
		return nil
	}
	return containers
}

func convertCharmMounts(input []params.CharmMount) []charm.Mount {
	mounts := []charm.Mount(nil)
	for _, v := range input {
		mounts = append(mounts, charm.Mount{
			Storage:  v.Storage,
			Location: v.Location,
		})
	}
	return mounts
}
