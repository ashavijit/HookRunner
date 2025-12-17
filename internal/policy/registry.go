package policy

import (
	"fmt"
	"regexp"
)

type Registry struct {
	fetcher *Fetcher
	workDir string
}

func NewRegistry(workDir, cacheDir string) *Registry {
	return &Registry{
		fetcher: NewFetcher(cacheDir),
		workDir: workDir,
	}
}

func (r *Registry) Load(config *UserConfig) (*MergedPolicy, error) {
	if config == nil {
		return nil, nil
	}

	merged := &MergedPolicy{
		LocalPolicies: config.LocalPolicies,
	}

	for _, ref := range config.Policies {
		policy, err := r.fetcher.LoadPolicy(ref.URL)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", ref.URL, err)
		}
		merged.RemotePolicies = append(merged.RemotePolicies, *policy)
	}

	merged.EffectiveRules = r.resolveRules(merged)
	return merged, nil
}

func (r *Registry) resolveRules(merged *MergedPolicy) PolicyRules {
	var effective PolicyRules

	for _, remote := range merged.RemotePolicies {
		effective = effective.Merge(remote.Rules)
	}

	for _, local := range merged.LocalPolicies {
		effective = effective.Merge(local.Rules)
	}

	return effective
}

func (r *Registry) Refresh(config *UserConfig) error {
	if err := r.fetcher.ClearCache(); err != nil {
		return err
	}
	_, err := r.Load(config)
	return err
}

func (r *Registry) ClearCache() error {
	return r.fetcher.ClearCache()
}

func ValidatePolicy(p *RemotePolicy) error {
	if p.Name == "" && p.ID != "" {
		p.Name = p.ID
	}

	if p.Name == "" {
		return fmt.Errorf("policy name required")
	}

	if p.Version != "" && p.Version != "local" {
		versionPattern := regexp.MustCompile(`^\d+(\.\d+)*$`)
		if !versionPattern.MatchString(p.Version) {
			return fmt.Errorf("version must be semver or 'local': %s", p.Version)
		}
	}

	return nil
}

func ValidateLocalPolicy(p *LocalPolicy) error {
	if p.Name == "" {
		return fmt.Errorf("policy name required")
	}

	if p.Version == "" {
		p.Version = "local"
	}

	return nil
}
