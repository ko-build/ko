//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// EnvironmentsService handles communication with the environment related methods
// of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/environments.html
type EnvironmentsService struct {
	client *Client
}

// Environment represents a GitLab environment.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/environments.html
type Environment struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Slug           string      `json:"slug"`
	State          string      `json:"state"`
	Tier           string      `json:"tier"`
	ExternalURL    string      `json:"external_url"`
	Project        *Project    `json:"project"`
	CreatedAt      *time.Time  `json:"created_at"`
	UpdatedAt      *time.Time  `json:"updated_at"`
	LastDeployment *Deployment `json:"last_deployment"`
}

func (env Environment) String() string {
	return Stringify(env)
}

// ListEnvironmentsOptions represents the available ListEnvironments() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#list-environments
type ListEnvironmentsOptions struct {
	ListOptions
	Name   *string `url:"name,omitempty" json:"name,omitempty"`
	Search *string `url:"search,omitempty" json:"search,omitempty"`
	States *string `url:"states,omitempty" json:"states,omitempty"`
}

// ListEnvironments gets a list of environments from a project, sorted by name
// alphabetically.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#list-environments
func (s *EnvironmentsService) ListEnvironments(pid interface{}, opts *ListEnvironmentsOptions, options ...RequestOptionFunc) ([]*Environment, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/environments", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var envs []*Environment
	resp, err := s.client.Do(req, &envs)
	if err != nil {
		return nil, resp, err
	}

	return envs, resp, err
}

// GetEnvironment gets a specific environment from a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#get-a-specific-environment
func (s *EnvironmentsService) GetEnvironment(pid interface{}, environment int, options ...RequestOptionFunc) (*Environment, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/environments/%d", PathEscape(project), environment)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	env := new(Environment)
	resp, err := s.client.Do(req, env)
	if err != nil {
		return nil, resp, err
	}

	return env, resp, err
}

// CreateEnvironmentOptions represents the available CreateEnvironment() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#create-a-new-environment
type CreateEnvironmentOptions struct {
	Name        *string `url:"name,omitempty" json:"name,omitempty"`
	ExternalURL *string `url:"external_url,omitempty" json:"external_url,omitempty"`
	Tier        *string `url:"tier,omitempty" json:"tier,omitempty"`
}

// CreateEnvironment adds an environment to a project. This is an idempotent
// method and can be called multiple times with the same parameters. Createing
// an environment that is already a environment does not affect the
// existing environmentship.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#create-a-new-environment
func (s *EnvironmentsService) CreateEnvironment(pid interface{}, opt *CreateEnvironmentOptions, options ...RequestOptionFunc) (*Environment, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/environments", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	env := new(Environment)
	resp, err := s.client.Do(req, env)
	if err != nil {
		return nil, resp, err
	}

	return env, resp, err
}

// EditEnvironmentOptions represents the available EditEnvironment() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#edit-an-existing-environment
type EditEnvironmentOptions struct {
	Name        *string `url:"name,omitempty" json:"name,omitempty"`
	ExternalURL *string `url:"external_url,omitempty" json:"external_url,omitempty"`
	Tier        *string `url:"tier,omitempty" json:"tier,omitempty"`
}

// EditEnvironment updates a project team environment to a specified access level..
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/environments.html#edit-an-existing-environment
func (s *EnvironmentsService) EditEnvironment(pid interface{}, environment int, opt *EditEnvironmentOptions, options ...RequestOptionFunc) (*Environment, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/environments/%d", PathEscape(project), environment)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	env := new(Environment)
	resp, err := s.client.Do(req, env)
	if err != nil {
		return nil, resp, err
	}

	return env, resp, err
}

// DeleteEnvironment removes an environment from a project team.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/environments.html#remove-a-environment-from-a-group-or-project
func (s *EnvironmentsService) DeleteEnvironment(pid interface{}, environment int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/environments/%d", PathEscape(project), environment)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// StopEnvironment stop an environment from a project team.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/environments.html#stop-an-environment
func (s *EnvironmentsService) StopEnvironment(pid interface{}, environmentID int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/environments/%d/stop", PathEscape(project), environmentID)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
