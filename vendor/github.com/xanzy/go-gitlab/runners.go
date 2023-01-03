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

// RunnersService handles communication with the runner related methods of the
// GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/runners.html
type RunnersService struct {
	client *Client
}

// Runner represents a GitLab CI Runner.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/runners.html
type Runner struct {
	ID             int        `json:"id"`
	Description    string     `json:"description"`
	Active         bool       `json:"active"`
	Paused         bool       `json:"paused"`
	IsShared       bool       `json:"is_shared"`
	IPAddress      string     `json:"ip_address"`
	RunnerType     string     `json:"runner_type"`
	Name           string     `json:"name"`
	Online         bool       `json:"online"`
	Status         string     `json:"status"`
	Token          string     `json:"token"`
	TokenExpiresAt *time.Time `json:"token_expires_at"`
}

// RunnerDetails represents the GitLab CI runner details.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/runners.html
type RunnerDetails struct {
	Paused       bool       `json:"paused"`
	Architecture string     `json:"architecture"`
	Description  string     `json:"description"`
	ID           int        `json:"id"`
	IPAddress    string     `json:"ip_address"`
	IsShared     bool       `json:"is_shared"`
	RunnerType   string     `json:"runner_type"`
	ContactedAt  *time.Time `json:"contacted_at"`
	Name         string     `json:"name"`
	Online       bool       `json:"online"`
	Status       string     `json:"status"`
	Platform     string     `json:"platform"`
	Projects     []struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		NameWithNamespace string `json:"name_with_namespace"`
		Path              string `json:"path"`
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"projects"`
	Token          string   `json:"token"`
	Revision       string   `json:"revision"`
	TagList        []string `json:"tag_list"`
	RunUntagged    bool     `json:"run_untagged"`
	Version        string   `json:"version"`
	Locked         bool     `json:"locked"`
	AccessLevel    string   `json:"access_level"`
	MaximumTimeout int      `json:"maximum_timeout"`
	Groups         []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		WebURL string `json:"web_url"`
	} `json:"groups"`

	// Deprecated members
	Active bool `json:"active"`
}

// ListRunnersOptions represents the available ListRunners() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-owned-runners
type ListRunnersOptions struct {
	ListOptions
	Type    *string   `url:"type,omitempty" json:"type,omitempty"`
	Status  *string   `url:"status,omitempty" json:"status,omitempty"`
	Paused  *bool     `url:"paused,omitempty" json:"paused,omitempty"`
	TagList *[]string `url:"tag_list,comma,omitempty" json:"tag_list,omitempty"`

	// Deprecated members
	Scope *string `url:"scope,omitempty" json:"scope,omitempty"`
}

// ListRunners gets a list of runners accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-owned-runners
func (s *RunnersService) ListRunners(opt *ListRunnersOptions, options ...RequestOptionFunc) ([]*Runner, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "runners", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var rs []*Runner
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// ListAllRunners gets a list of all runners in the GitLab instance. Access is
// restricted to users with admin privileges.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-all-runners
func (s *RunnersService) ListAllRunners(opt *ListRunnersOptions, options ...RequestOptionFunc) ([]*Runner, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "runners/all", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var rs []*Runner
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// GetRunnerDetails returns details for given runner.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#get-runner-39-s-details
func (s *RunnersService) GetRunnerDetails(rid interface{}, options ...RequestOptionFunc) (*RunnerDetails, *Response, error) {
	runner, err := parseID(rid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("runners/%s", runner)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	rs := new(RunnerDetails)
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// UpdateRunnerDetailsOptions represents the available UpdateRunnerDetails() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#update-runner-39-s-details
type UpdateRunnerDetailsOptions struct {
	Description    *string   `url:"description,omitempty" json:"description,omitempty"`
	Paused         *bool     `url:"paused,omitempty" json:"paused,omitempty"`
	TagList        *[]string `url:"tag_list[],omitempty" json:"tag_list,omitempty"`
	RunUntagged    *bool     `url:"run_untagged,omitempty" json:"run_untagged,omitempty"`
	Locked         *bool     `url:"locked,omitempty" json:"locked,omitempty"`
	AccessLevel    *string   `url:"access_level,omitempty" json:"access_level,omitempty"`
	MaximumTimeout *int      `url:"maximum_timeout,omitempty" json:"maximum_timeout,omitempty"`

	// Deprecated members
	Active *bool `url:"active,omitempty" json:"active,omitempty"`
}

// UpdateRunnerDetails updates details for a given runner.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#update-runner-39-s-details
func (s *RunnersService) UpdateRunnerDetails(rid interface{}, opt *UpdateRunnerDetailsOptions, options ...RequestOptionFunc) (*RunnerDetails, *Response, error) {
	runner, err := parseID(rid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("runners/%s", runner)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	rs := new(RunnerDetails)
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// RemoveRunner removes a runner.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#remove-a-runner
func (s *RunnersService) RemoveRunner(rid interface{}, options ...RequestOptionFunc) (*Response, error) {
	runner, err := parseID(rid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("runners/%s", runner)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListRunnerJobsOptions represents the available ListRunnerJobs()
// options. Status can be one of: running, success, failed, canceled.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-runners-jobs
type ListRunnerJobsOptions struct {
	ListOptions
	Status  *string `url:"status,omitempty" json:"status,omitempty"`
	OrderBy *string `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort    *string `url:"sort,omitempty" json:"sort,omitempty"`
}

// ListRunnerJobs gets a list of jobs that are being processed or were processed by specified Runner.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-runner-39-s-jobs
func (s *RunnersService) ListRunnerJobs(rid interface{}, opt *ListRunnerJobsOptions, options ...RequestOptionFunc) ([]*Job, *Response, error) {
	runner, err := parseID(rid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("runners/%s/jobs", runner)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var rs []*Job
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// ListProjectRunnersOptions represents the available ListProjectRunners()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-project-s-runners
type ListProjectRunnersOptions ListRunnersOptions

// ListProjectRunners gets a list of runners accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#list-project-s-runners
func (s *RunnersService) ListProjectRunners(pid interface{}, opt *ListProjectRunnersOptions, options ...RequestOptionFunc) ([]*Runner, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/runners", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var rs []*Runner
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// EnableProjectRunnerOptions represents the available EnableProjectRunner()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#enable-a-runner-in-project
type EnableProjectRunnerOptions struct {
	RunnerID int `json:"runner_id"`
}

// EnableProjectRunner enables an available specific runner in the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#enable-a-runner-in-project
func (s *RunnersService) EnableProjectRunner(pid interface{}, opt *EnableProjectRunnerOptions, options ...RequestOptionFunc) (*Runner, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/runners", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(Runner)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}

// DisableProjectRunner disables a specific runner from project.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#disable-a-runner-from-project
func (s *RunnersService) DisableProjectRunner(pid interface{}, runner int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/runners/%d", PathEscape(project), runner)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListGroupsRunnersOptions represents the available ListGroupsRunners() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#list-groups-runners
type ListGroupsRunnersOptions struct {
	ListOptions
	Type    *string   `url:"type,omitempty" json:"type,omitempty"`
	Status  *string   `url:"status,omitempty" json:"status,omitempty"`
	TagList *[]string `url:"tag_list,comma,omitempty" json:"tag_list,omitempty"`
}

// ListGroupsRunners lists all runners (specific and shared) available in the
// group as well it’s ancestor groups. Shared runners are listed if at least one
// shared runner is defined.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#list-groups-runners
func (s *RunnersService) ListGroupsRunners(gid interface{}, opt *ListGroupsRunnersOptions, options ...RequestOptionFunc) ([]*Runner, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/runners", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var rs []*Runner
	resp, err := s.client.Do(req, &rs)
	if err != nil {
		return nil, resp, err
	}

	return rs, resp, err
}

// RegisterNewRunnerOptions represents the available RegisterNewRunner()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#register-a-new-runner
type RegisterNewRunnerOptions struct {
	Token           *string                       `url:"token" json:"token"`
	Description     *string                       `url:"description,omitempty" json:"description,omitempty"`
	Info            *RegisterNewRunnerInfoOptions `url:"info,omitempty" json:"info,omitempty"`
	Active          *bool                         `url:"active,omitempty" json:"active,omitempty"`
	Paused          *bool                         `url:"paused,omitempty" json:"paused,omitempty"`
	Locked          *bool                         `url:"locked,omitempty" json:"locked,omitempty"`
	RunUntagged     *bool                         `url:"run_untagged,omitempty" json:"run_untagged,omitempty"`
	TagList         *[]string                     `url:"tag_list[],omitempty" json:"tag_list,omitempty"`
	AccessLevel     *string                       `url:"access_level,omitempty" json:"access_level,omitempty"`
	MaximumTimeout  *int                          `url:"maximum_timeout,omitempty" json:"maximum_timeout,omitempty"`
	MaintenanceNote *string                       `url:"maintenance_note,omitempty" json:"maintenance_note,omitempty"`
}

// RegisterNewRunnerInfoOptions represents the info hashmap parameter in
// RegisterNewRunnerOptions.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#register-a-new-runner
type RegisterNewRunnerInfoOptions struct {
	Name         *string `url:"name,omitempty" json:"name,omitempty"`
	Version      *string `url:"version,omitempty" json:"version,omitempty"`
	Revision     *string `url:"revision,omitempty" json:"revision,omitempty"`
	Platform     *string `url:"platform,omitempty" json:"platform,omitempty"`
	Architecture *string `url:"architecture,omitempty" json:"architecture,omitempty"`
}

// RegisterNewRunner registers a new Runner for the instance.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#register-a-new-runner
func (s *RunnersService) RegisterNewRunner(opt *RegisterNewRunnerOptions, options ...RequestOptionFunc) (*Runner, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "runners", opt, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(Runner)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}

// DeleteRegisteredRunnerOptions represents the available
// DeleteRegisteredRunner() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#delete-a-registered-runner
type DeleteRegisteredRunnerOptions struct {
	Token *string `url:"token" json:"token"`
}

// DeleteRegisteredRunner deletes a Runner by Token.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#delete-a-runner-by-authentication-token
func (s *RunnersService) DeleteRegisteredRunner(opt *DeleteRegisteredRunnerOptions, options ...RequestOptionFunc) (*Response, error) {
	req, err := s.client.NewRequest(http.MethodDelete, "runners", opt, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteRegisteredRunnerByID deletes a runner by ID.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#delete-a-runner-by-id
func (s *RunnersService) DeleteRegisteredRunnerByID(rid int, options ...RequestOptionFunc) (*Response, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("runners/%d", rid), nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// VerifyRegisteredRunnerOptions represents the available
// VerifyRegisteredRunner() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#verify-authentication-for-a-registered-runner
type VerifyRegisteredRunnerOptions struct {
	Token *string `url:"token" json:"token"`
}

// VerifyRegisteredRunner registers a new runner for the instance.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/runners.html#verify-authentication-for-a-registered-runner
func (s *RunnersService) VerifyRegisteredRunner(opt *VerifyRegisteredRunnerOptions, options ...RequestOptionFunc) (*Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "runners/verify", opt, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

type RunnerRegistrationToken struct {
	Token          *string    `url:"token" json:"token"`
	TokenExpiresAt *time.Time `url:"token_expires_at" json:"token_expires_at"`
}

// ResetInstanceRunnerRegistrationToken resets the instance runner registration
// token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#reset-instances-runner-registration-token
func (s *RunnersService) ResetInstanceRunnerRegistrationToken(options ...RequestOptionFunc) (*RunnerRegistrationToken, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "runners/reset_registration_token", nil, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(RunnerRegistrationToken)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}

// ResetGroupRunnerRegistrationToken resets a group's runner registration token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#reset-groups-runner-registration-token
func (s *RunnersService) ResetGroupRunnerRegistrationToken(gid interface{}, options ...RequestOptionFunc) (*RunnerRegistrationToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/runners/reset_registration_token", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(RunnerRegistrationToken)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}

// ResetGroupRunnerRegistrationToken resets a projects's runner registration token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#reset-projects-runner-registration-token
func (s *RunnersService) ResetProjectRunnerRegistrationToken(pid interface{}, options ...RequestOptionFunc) (*RunnerRegistrationToken, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/runners/reset_registration_token", PathEscape(project))
	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(RunnerRegistrationToken)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}

type RunnerAuthenticationToken struct {
	Token          *string    `url:"token" json:"token"`
	TokenExpiresAt *time.Time `url:"token_expires_at" json:"token_expires_at"`
}

// ResetRunnerAuthenticationToken resets a runner's authentication token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/runners.html#reset-runners-authentication-token
func (s *RunnersService) ResetRunnerAuthenticationToken(rid int, options ...RequestOptionFunc) (*RunnerAuthenticationToken, *Response, error) {
	u := fmt.Sprintf("runners/%d/reset_authentication_token", rid)
	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	r := new(RunnerAuthenticationToken)
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, err
}
