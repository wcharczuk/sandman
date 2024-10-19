package configmeta

import (
	"context"

	"go.charczuk.com/sdk/configutil"
)

// These are set with `-ldflags="-X` on `go install`
var (
	// Version is the current version.
	Version = ""
	// GitRef is the currently deployed git ref
	GitRef = "HEAD"
	// ServiceName is the name of the service
	ServiceName = ""
	// ProjectName is the name of the project the service belongs to
	ProjectName = ""
	// ClusterName is the name of the cluster the service is deployed to
	ClusterName = ""
	// Region is the region the service is deployed to
	Region = ""
)

const (
	// DefaultVersion is the default version.
	DefaultVersion = "HEAD"
)

// Meta is the cluster config meta.
type Meta struct {
	// Region is the aws region the service is deployed to.
	Region string `json:"region,omitempty" yaml:"region,omitempty"`
	// ServiceName is name of the service
	ServiceName string `json:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	// ProjectName is the project name injected by Deployinator.
	ProjectName string `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	// ClusterName is the name of the cluster the service is deployed to
	ClusterName string `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	// Environment is the environment of the cluster (sandbox, prod etc.)
	ServiceEnv string `json:"serviceEnv,omitempty" yaml:"serviceEnv,omitempty"`
	// Hostname is the environment of the cluster (sandbox, prod etc.)
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// Version is the application version.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// GitRef is the git ref of the image.
	GitRef string `json:"gitRef,omitempty" yaml:"gitRef,omitempty"`
}

// SetFrom returns a resolve action to set this meta from a root meta.
func (m *Meta) SetFrom(other *Meta) configutil.ResolveAction {
	return func(_ context.Context) error {
		m.Region = other.Region
		m.ServiceName = other.ServiceName
		m.ProjectName = other.ProjectName
		m.ClusterName = other.ClusterName
		m.ServiceEnv = other.ServiceEnv
		m.Hostname = other.Hostname
		m.Version = other.Version
		m.GitRef = other.GitRef
		return nil
	}
}

// ApplyTo applies a given meta to another meta.
func (m *Meta) ApplyTo(other *Meta) configutil.ResolveAction {
	return func(_ context.Context) error {
		other.Region = m.Region
		other.ServiceName = m.ServiceName
		other.ProjectName = m.ProjectName
		other.ClusterName = m.ClusterName
		other.ServiceEnv = m.ServiceEnv
		other.Hostname = m.Hostname
		other.Version = m.Version
		other.GitRef = m.GitRef
		return nil
	}
}

// Resolve implements configutil.Resolver
func (m *Meta) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&m.Region, configutil.Env[string](VarRegion), configutil.Lazy(&m.Region), configutil.Lazy(&Region)),
		configutil.Set(&m.ServiceName, configutil.Env[string](VarServiceName), configutil.Lazy(&m.ServiceName), configutil.Lazy(&ServiceName)),
		configutil.Set(&m.ProjectName, configutil.Env[string](VarProjectName), configutil.Lazy(&m.ProjectName), configutil.Lazy(&ProjectName)),
		configutil.Set(&m.ClusterName, configutil.Env[string](VarClusterName), configutil.Lazy(&m.ClusterName), configutil.Lazy(&ClusterName)),
		configutil.Set(&m.ServiceEnv, configutil.Env[string](VarServiceEnv), configutil.Lazy(&m.ServiceEnv)),
		configutil.Set(&m.Hostname, configutil.Env[string](VarHostname), configutil.Lazy(&m.Hostname)),
		configutil.Set(&m.Version, configutil.Env[string](VarVersion), configutil.Lazy(&m.Version), configutil.Lazy(&Version), configutil.Const(DefaultVersion)),
		configutil.Set(&m.GitRef, configutil.Env[string](VarGitRef), configutil.Lazy(&m.GitRef), configutil.Lazy(&GitRef)),
	)
}

// IsProdlike returns if the ServiceEnv is prodlike, that is
// an environment where we care about secrets leaking.
func (m Meta) IsProdlike() bool {
	return m.ServiceEnv != EnvDev && m.ServiceEnv != EnvTest
}
