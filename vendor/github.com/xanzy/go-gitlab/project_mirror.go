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

// ProjectMirrorService handles communication with the project mirror
// related methods of the GitLab API.
//
// GitLAb API docs: https://docs.gitlab.com/ce/api/remote_mirrors.html
type ProjectMirrorService struct {
	client *Client
}

// ProjectMirror represents a project mirror configuration.
//
// GitLAb API docs: https://docs.gitlab.com/ce/api/remote_mirrors.html
type ProjectMirror struct {
	Enabled                bool       `json:"enabled"`
	ID                     int        `json:"id"`
	LastError              string     `json:"last_error"`
	LastSuccessfulUpdateAt *time.Time `json:"last_successful_update_at"`
	LastUpdateAt           *time.Time `json:"last_update_at"`
	LastUpdateStartedAt    *time.Time `json:"last_update_started_at"`
	OnlyProtectedBranches  bool       `json:"only_protected_branches"`
	KeepDivergentRefs      bool       `json:"keep_divergent_refs"`
	UpdateStatus           string     `json:"update_status"`
	URL                    string     `json:"url"`
}

// ListProjectMirrorOptions represents the available ListProjectMirror() options.
type ListProjectMirrorOptions ListOptions

// ListProjectMirror gets a list of mirrors configured on the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/remote_mirrors.html#list-a-projects-remote-mirrors
func (s *ProjectMirrorService) ListProjectMirror(pid interface{}, opt *ListProjectMirrorOptions, options ...RequestOptionFunc) ([]*ProjectMirror, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/remote_mirrors", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pm []*ProjectMirror
	resp, err := s.client.Do(req, &pm)
	if err != nil {
		return nil, resp, err
	}

	return pm, resp, err
}

// GetProjectMirror gets a single mirror configured on the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/remote_mirrors.html#get-a-single-projects-remote-mirror
func (s *ProjectMirrorService) GetProjectMirror(pid interface{}, mirror int, options ...RequestOptionFunc) (*ProjectMirror, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/remote_mirrors/%d", PathEscape(project), mirror)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pm := new(ProjectMirror)
	resp, err := s.client.Do(req, &pm)
	if err != nil {
		return nil, resp, err
	}

	return pm, resp, err
}

// AddProjectMirrorOptions contains the properties requires to create
// a new project mirror.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/remote_mirrors.html#create-a-remote-mirror
type AddProjectMirrorOptions struct {
	URL                   *string `url:"url,omitempty" json:"url,omitempty"`
	Enabled               *bool   `url:"enabled,omitempty" json:"enabled,omitempty"`
	OnlyProtectedBranches *bool   `url:"only_protected_branches,omitempty" json:"only_protected_branches,omitempty"`
	KeepDivergentRefs     *bool   `url:"keep_divergent_refs,omitempty" json:"keep_divergent_refs,omitempty"`
}

// AddProjectMirror creates a new mirror on the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/remote_mirrors.html#create-a-remote-mirror
func (s *ProjectMirrorService) AddProjectMirror(pid interface{}, opt *AddProjectMirrorOptions, options ...RequestOptionFunc) (*ProjectMirror, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/remote_mirrors", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pm := new(ProjectMirror)
	resp, err := s.client.Do(req, pm)
	if err != nil {
		return nil, resp, err
	}

	return pm, resp, err
}

// EditProjectMirrorOptions contains the properties requires to edit
// an existing project mirror.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/remote_mirrors.html#update-a-remote-mirrors-attributes
type EditProjectMirrorOptions struct {
	Enabled               *bool `url:"enabled,omitempty" json:"enabled,omitempty"`
	OnlyProtectedBranches *bool `url:"only_protected_branches,omitempty" json:"only_protected_branches,omitempty"`
	KeepDivergentRefs     *bool `url:"keep_divergent_refs,omitempty" json:"keep_divergent_refs,omitempty"`
}

// EditProjectMirror updates a project team member to a specified access level..
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/remote_mirrors.html#update-a-remote-mirrors-attributes
func (s *ProjectMirrorService) EditProjectMirror(pid interface{}, mirror int, opt *EditProjectMirrorOptions, options ...RequestOptionFunc) (*ProjectMirror, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/remote_mirrors/%d", PathEscape(project), mirror)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pm := new(ProjectMirror)
	resp, err := s.client.Do(req, pm)
	if err != nil {
		return nil, resp, err
	}

	return pm, resp, err
}

// DeleteProjectMirror deletes a project mirror.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/remote_mirrors.html#delete-a-remote-mirror
func (s *ProjectMirrorService) DeleteProjectMirror(pid interface{}, mirror int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/remote_mirrors/%d", PathEscape(project), mirror)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
