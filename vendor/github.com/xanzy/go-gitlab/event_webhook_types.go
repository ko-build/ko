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
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// BuildEvent represents a build event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#build-events
type BuildEvent struct {
	ObjectKind        string     `json:"object_kind"`
	Ref               string     `json:"ref"`
	Tag               bool       `json:"tag"`
	BeforeSHA         string     `json:"before_sha"`
	SHA               string     `json:"sha"`
	BuildID           int        `json:"build_id"`
	BuildName         string     `json:"build_name"`
	BuildStage        string     `json:"build_stage"`
	BuildStatus       string     `json:"build_status"`
	BuildCreatedAt    string     `json:"build_created_at"`
	BuildStartedAt    string     `json:"build_started_at"`
	BuildFinishedAt   string     `json:"build_finished_at"`
	BuildDuration     float64    `json:"build_duration"`
	BuildAllowFailure bool       `json:"build_allow_failure"`
	ProjectID         int        `json:"project_id"`
	ProjectName       string     `json:"project_name"`
	User              *EventUser `json:"user"`
	Commit            struct {
		ID          int    `json:"id"`
		SHA         string `json:"sha"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
		Status      string `json:"status"`
		Duration    int    `json:"duration"`
		StartedAt   string `json:"started_at"`
		FinishedAt  string `json:"finished_at"`
	} `json:"commit"`
	Repository *Repository `json:"repository"`
}

// CommitCommentEvent represents a comment on a commit event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#comment-on-a-commit
type CommitCommentEvent struct {
	ObjectKind string `json:"object_kind"`
	User       *User  `json:"user"`
	ProjectID  int    `json:"project_id"`
	Project    struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID           int    `json:"id"`
		Note         string `json:"note"`
		NoteableType string `json:"noteable_type"`
		AuthorID     int    `json:"author_id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		ProjectID    int    `json:"project_id"`
		Attachment   string `json:"attachment"`
		LineCode     string `json:"line_code"`
		CommitID     string `json:"commit_id"`
		NoteableID   int    `json:"noteable_id"`
		System       bool   `json:"system"`
		StDiff       struct {
			Diff        string `json:"diff"`
			NewPath     string `json:"new_path"`
			OldPath     string `json:"old_path"`
			AMode       string `json:"a_mode"`
			BMode       string `json:"b_mode"`
			NewFile     bool   `json:"new_file"`
			RenamedFile bool   `json:"renamed_file"`
			DeletedFile bool   `json:"deleted_file"`
		} `json:"st_diff"`
		Description string `json:"description"`
		URL         string `json:"url"`
	} `json:"object_attributes"`
	Commit *struct {
		ID        string     `json:"id"`
		Title     string     `json:"title"`
		Message   string     `json:"message"`
		Timestamp *time.Time `json:"timestamp"`
		URL       string     `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commit"`
}

// DeploymentEvent represents a deployment event
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#deployment-events
type DeploymentEvent struct {
	ObjectKind    string `json:"object_kind"`
	Status        string `json:"status"`
	DeployableID  int    `json:"deployable_id"`
	DeployableURL string `json:"deployable_url"`
	Environment   string `json:"environment"`
	Project       struct {
		ID                int     `json:"id"`
		Name              string  `json:"name"`
		Description       string  `json:"description"`
		WebURL            string  `json:"web_url"`
		AvatarURL         *string `json:"avatar_url"`
		GitSSHURL         string  `json:"git_ssh_url"`
		GitHTTPURL        string  `json:"git_http_url"`
		Namespace         string  `json:"namespace"`
		VisibilityLevel   int     `json:"visibility_level"`
		PathWithNamespace string  `json:"path_with_namespace"`
		DefaultBranch     string  `json:"default_branch"`
		CIConfigPath      string  `json:"ci_config_path"`
		Homepage          string  `json:"homepage"`
		URL               string  `json:"url"`
		SSHURL            string  `json:"ssh_url"`
		HTTPURL           string  `json:"http_url"`
	} `json:"project"`
	Ref         string     `json:"ref"`
	ShortSHA    string     `json:"short_sha"`
	User        *EventUser `json:"user"`
	UserURL     string     `json:"user_url"`
	CommitURL   string     `json:"commit_url"`
	CommitTitle string     `json:"commit_title"`
}

// IssueCommentEvent represents a comment on an issue event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#comment-on-an-issue
type IssueCommentEvent struct {
	ObjectKind string `json:"object_kind"`
	User       *User  `json:"user"`
	ProjectID  int    `json:"project_id"`
	Project    struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID           int     `json:"id"`
		Note         string  `json:"note"`
		NoteableType string  `json:"noteable_type"`
		AuthorID     int     `json:"author_id"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
		ProjectID    int     `json:"project_id"`
		Attachment   string  `json:"attachment"`
		LineCode     string  `json:"line_code"`
		CommitID     string  `json:"commit_id"`
		DiscussionID string  `json:"discussion_id"`
		NoteableID   int     `json:"noteable_id"`
		System       bool    `json:"system"`
		StDiff       []*Diff `json:"st_diff"`
		Description  string  `json:"description"`
		URL          string  `json:"url"`
	} `json:"object_attributes"`
	Issue struct {
		ID                  int           `json:"id"`
		IID                 int           `json:"iid"`
		ProjectID           int           `json:"project_id"`
		MilestoneID         int           `json:"milestone_id"`
		AuthorID            int           `json:"author_id"`
		Description         string        `json:"description"`
		State               string        `json:"state"`
		Title               string        `json:"title"`
		Labels              []*EventLabel `json:"labels"`
		LastEditedAt        string        `json:"last_edit_at"`
		LastEditedByID      int           `json:"last_edited_by_id"`
		UpdatedAt           string        `json:"updated_at"`
		UpdatedByID         int           `json:"updated_by_id"`
		CreatedAt           string        `json:"created_at"`
		ClosedAt            string        `json:"closed_at"`
		DueDate             *ISOTime      `json:"due_date"`
		URL                 string        `json:"url"`
		TimeEstimate        int           `json:"time_estimate"`
		Confidential        bool          `json:"confidential"`
		TotalTimeSpent      int           `json:"total_time_spent"`
		HumanTotalTimeSpent string        `json:"human_total_time_spent"`
		HumanTimeEstimate   string        `json:"human_time_estimate"`
		AssigneeIDs         []int         `json:"assignee_ids"`
		AssigneeID          int           `json:"assignee_id"`
	} `json:"issue"`
}

// IssueEvent represents a issue event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#issue-events
type IssueEvent struct {
	ObjectKind string     `json:"object_kind"`
	User       *EventUser `json:"user"`
	Project    struct {
		ID                int             `json:"id"`
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		AssigneeID  int    `json:"assignee_id"`
		AuthorID    int    `json:"author_id"`
		ProjectID   int    `json:"project_id"`
		CreatedAt   string `json:"created_at"` // Should be *time.Time (see Gitlab issue #21468)
		UpdatedAt   string `json:"updated_at"` // Should be *time.Time (see Gitlab issue #21468)
		Position    int    `json:"position"`
		BranchName  string `json:"branch_name"`
		Description string `json:"description"`
		MilestoneID int    `json:"milestone_id"`
		State       string `json:"state"`
		IID         int    `json:"iid"`
		URL         string `json:"url"`
		Action      string `json:"action"`
	} `json:"object_attributes"`
	Assignee  *EventUser    `json:"assignee"`
	Assignees *[]EventUser  `json:"assignees"`
	Labels    []*EventLabel `json:"labels"`
	Changes   struct {
		Assignees struct {
			Previous []*EventUser `json:"previous"`
			Current  []*EventUser `json:"current"`
		} `json:"assignees"`
		Description struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"description"`
		Labels struct {
			Previous []*EventLabel `json:"previous"`
			Current  []*EventLabel `json:"current"`
		} `json:"labels"`
		Title struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"title"`
		UpdatedByID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"updated_by_id"`
		TotalTimeSpent struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"total_time_spent"`
	} `json:"changes"`
}

// JobEvent represents a job event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#job-events
type JobEvent struct {
	ObjectKind         string     `json:"object_kind"`
	Ref                string     `json:"ref"`
	Tag                bool       `json:"tag"`
	BeforeSHA          string     `json:"before_sha"`
	SHA                string     `json:"sha"`
	BuildID            int        `json:"build_id"`
	BuildName          string     `json:"build_name"`
	BuildStage         string     `json:"build_stage"`
	BuildStatus        string     `json:"build_status"`
	BuildStartedAt     string     `json:"build_started_at"`
	BuildFinishedAt    string     `json:"build_finished_at"`
	BuildDuration      float64    `json:"build_duration"`
	BuildAllowFailure  bool       `json:"build_allow_failure"`
	BuildFailureReason string     `json:"build_failure_reason"`
	PipelineID         int        `json:"pipeline_id"`
	ProjectID          int        `json:"project_id"`
	ProjectName        string     `json:"project_name"`
	User               *EventUser `json:"user"`
	Commit             struct {
		ID          int    `json:"id"`
		SHA         string `json:"sha"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
		AuthorURL   string `json:"author_url"`
		Status      string `json:"status"`
		Duration    int    `json:"duration"`
		StartedAt   string `json:"started_at"`
		FinishedAt  string `json:"finished_at"`
	} `json:"commit"`
	Repository *Repository `json:"repository"`
	Runner     struct {
		ID          int    `json:"id"`
		Active      bool   `json:"active"`
		Shared      bool   `json:"is_shared"`
		Description string `json:"description"`
	} `json:"runner"`
}

// MemberEvent represents a member event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#group-member-events
type MemberEvent struct {
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	GroupName    string     `json:"group_name"`
	GroupPath    string     `json:"group_path"`
	GroupID      int        `json:"group_id"`
	UserUsername string     `json:"user_username"`
	UserName     string     `json:"user_name"`
	UserEmail    string     `json:"user_email"`
	UserID       int        `json:"user_id"`
	GroupAccess  string     `json:"group_access"`
	GroupPlan    string     `json:"group_plan"`
	ExpiresAt    *time.Time `json:"expires_at"`
	EventName    string     `json:"event_name"`
}

// MergeCommentEvent represents a comment on a merge event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#comment-on-a-merge-request
type MergeCommentEvent struct {
	ObjectKind string     `json:"object_kind"`
	User       *EventUser `json:"user"`
	ProjectID  int        `json:"project_id"`
	Project    struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	ObjectAttributes struct {
		Attachment       string        `json:"attachment"`
		AuthorID         int           `json:"author_id"`
		ChangePosition   *NotePosition `json:"change_position"`
		CommitID         string        `json:"commit_id"`
		CreatedAt        string        `json:"created_at"`
		DiscussionID     string        `json:"discussion_id"`
		ID               int           `json:"id"`
		LineCode         string        `json:"line_code"`
		Note             string        `json:"note"`
		NoteableID       int           `json:"noteable_id"`
		NoteableType     string        `json:"noteable_type"`
		OriginalPosition *NotePosition `json:"original_position"`
		Position         *NotePosition `json:"position"`
		ProjectID        int           `json:"project_id"`
		ResolvedAt       string        `json:"resolved_at"`
		ResolvedByID     int           `json:"resolved_by_id"`
		ResolvedByPush   bool          `json:"resolved_by_push"`
		StDiff           *Diff         `json:"st_diff"`
		System           bool          `json:"system"`
		Type             string        `json:"type"`
		UpdatedAt        string        `json:"updated_at"`
		UpdatedByID      string        `json:"updated_by_id"`
		Description      string        `json:"description"`
		URL              string        `json:"url"`
	} `json:"object_attributes"`
	Repository   *Repository `json:"repository"`
	MergeRequest struct {
		ID                        int          `json:"id"`
		TargetBranch              string       `json:"target_branch"`
		SourceBranch              string       `json:"source_branch"`
		SourceProjectID           int          `json:"source_project_id"`
		AuthorID                  int          `json:"author_id"`
		AssigneeID                int          `json:"assignee_id"`
		AssigneeIDs               []int        `json:"assignee_ids"`
		Title                     string       `json:"title"`
		CreatedAt                 string       `json:"created_at"`
		UpdatedAt                 string       `json:"updated_at"`
		MilestoneID               int          `json:"milestone_id"`
		State                     string       `json:"state"`
		MergeStatus               string       `json:"merge_status"`
		TargetProjectID           int          `json:"target_project_id"`
		IID                       int          `json:"iid"`
		Description               string       `json:"description"`
		Position                  int          `json:"position"`
		LockedAt                  string       `json:"locked_at"`
		UpdatedByID               int          `json:"updated_by_id"`
		MergeError                string       `json:"merge_error"`
		MergeParams               *MergeParams `json:"merge_params"`
		MergeWhenPipelineSucceeds bool         `json:"merge_when_pipeline_succeeds"`
		MergeUserID               int          `json:"merge_user_id"`
		MergeCommitSHA            string       `json:"merge_commit_sha"`
		DeletedAt                 string       `json:"deleted_at"`
		InProgressMergeCommitSHA  string       `json:"in_progress_merge_commit_sha"`
		LockVersion               int          `json:"lock_version"`
		ApprovalsBeforeMerge      string       `json:"approvals_before_merge"`
		RebaseCommitSHA           string       `json:"rebase_commit_sha"`
		TimeEstimate              int          `json:"time_estimate"`
		Squash                    bool         `json:"squash"`
		LastEditedAt              string       `json:"last_edited_at"`
		LastEditedByID            int          `json:"last_edited_by_id"`
		Source                    *Repository  `json:"source"`
		Target                    *Repository  `json:"target"`
		LastCommit                struct {
			ID        string     `json:"id"`
			Title     string     `json:"title"`
			Message   string     `json:"message"`
			Timestamp *time.Time `json:"timestamp"`
			URL       string     `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"last_commit"`
		WorkInProgress bool `json:"work_in_progress"`
		TotalTimeSpent int  `json:"total_time_spent"`
		HeadPipelineID int  `json:"head_pipeline_id"`
	} `json:"merge_request"`
}

// MergeEvent represents a merge event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#merge-request-events
type MergeEvent struct {
	ObjectKind string     `json:"object_kind"`
	User       *EventUser `json:"user"`
	Project    struct {
		ID                int             `json:"id"`
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	ObjectAttributes struct {
		ID                       int          `json:"id"`
		TargetBranch             string       `json:"target_branch"`
		SourceBranch             string       `json:"source_branch"`
		SourceProjectID          int          `json:"source_project_id"`
		AuthorID                 int          `json:"author_id"`
		AssigneeID               int          `json:"assignee_id"`
		AssigneeIDs              []int        `json:"assignee_ids"`
		ReviewerIDs              []int        `json:"reviewer_ids"`
		Title                    string       `json:"title"`
		CreatedAt                string       `json:"created_at"` // Should be *time.Time (see Gitlab issue #21468)
		UpdatedAt                string       `json:"updated_at"` // Should be *time.Time (see Gitlab issue #21468)
		StCommits                []*Commit    `json:"st_commits"`
		StDiffs                  []*Diff      `json:"st_diffs"`
		MilestoneID              int          `json:"milestone_id"`
		State                    string       `json:"state"`
		MergeStatus              string       `json:"merge_status"`
		TargetProjectID          int          `json:"target_project_id"`
		IID                      int          `json:"iid"`
		Description              string       `json:"description"`
		Position                 int          `json:"position"`
		LockedAt                 string       `json:"locked_at"`
		UpdatedByID              int          `json:"updated_by_id"`
		MergeError               string       `json:"merge_error"`
		MergeParams              *MergeParams `json:"merge_params"`
		MergeWhenBuildSucceeds   bool         `json:"merge_when_build_succeeds"`
		MergeUserID              int          `json:"merge_user_id"`
		MergeCommitSHA           string       `json:"merge_commit_sha"`
		DeletedAt                string       `json:"deleted_at"`
		ApprovalsBeforeMerge     string       `json:"approvals_before_merge"`
		RebaseCommitSHA          string       `json:"rebase_commit_sha"`
		InProgressMergeCommitSHA string       `json:"in_progress_merge_commit_sha"`
		LockVersion              int          `json:"lock_version"`
		TimeEstimate             int          `json:"time_estimate"`
		Source                   *Repository  `json:"source"`
		Target                   *Repository  `json:"target"`
		HeadPipelineID           *int         `json:"head_pipeline_id"`
		LastCommit               struct {
			ID        string     `json:"id"`
			Message   string     `json:"message"`
			Timestamp *time.Time `json:"timestamp"`
			URL       string     `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"last_commit"`
		BlockingDiscussionsResolved bool       `json:"blocking_discussions_resolved"`
		WorkInProgress              bool       `json:"work_in_progress"`
		URL                         string     `json:"url"`
		Action                      string     `json:"action"`
		OldRev                      string     `json:"oldrev"`
		Assignee                    *EventUser `json:"assignee"`
	} `json:"object_attributes"`
	Repository *Repository   `json:"repository"`
	Assignee   *EventUser    `json:"assignee"`
	Assignees  []*EventUser  `json:"assignees"`
	Reviewers  []*EventUser  `json:"reviewers"`
	Labels     []*EventLabel `json:"labels"`
	Changes    struct {
		Assignees struct {
			Previous []*EventUser `json:"previous"`
			Current  []*EventUser `json:"current"`
		} `json:"assignees"`
		Reviewers struct {
			Previous []*EventUser `json:"previous"`
			Current  []*EventUser `json:"current"`
		} `json:"reviewers"`
		Description struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"description"`
		Labels struct {
			Previous []*EventLabel `json:"previous"`
			Current  []*EventLabel `json:"current"`
		} `json:"labels"`
		SourceBranch struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"source_branch"`
		SourceProjectID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"source_project_id"`
		StateID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"state_id"`
		TargetBranch struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"target_branch"`
		TargetProjectID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"target_project_id"`
		Title struct {
			Previous string `json:"previous"`
			Current  string `json:"current"`
		} `json:"title"`
		UpdatedByID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"updated_by_id"`
		MilestoneID struct {
			Previous int `json:"previous"`
			Current  int `json:"current"`
		} `json:"milestone_id"`
	} `json:"changes"`
}

// EventUser represents a user record in an event and is used as an even initiator or a merge assignee.
type EventUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// MergeParams represents the merge params.
type MergeParams struct {
	ForceRemoveSourceBranch bool `json:"force_remove_source_branch"`
}

// UnmarshalJSON decodes the merge parameters
//
// This allows support of ForceRemoveSourceBranch for both type bool (>11.9) and string (<11.9)
func (p *MergeParams) UnmarshalJSON(b []byte) error {
	type Alias MergeParams
	raw := struct {
		*Alias
		ForceRemoveSourceBranch interface{} `json:"force_remove_source_branch"`
	}{
		Alias: (*Alias)(p),
	}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	switch v := raw.ForceRemoveSourceBranch.(type) {
	case nil:
		// No action needed.
	case bool:
		p.ForceRemoveSourceBranch = v
	case string:
		p.ForceRemoveSourceBranch, err = strconv.ParseBool(v)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("failed to unmarshal ForceRemoveSourceBranch of type: %T", v)
	}

	return nil
}

// PipelineEvent represents a pipeline event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#pipeline-events
type PipelineEvent struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes struct {
		ID             int      `json:"id"`
		Ref            string   `json:"ref"`
		Tag            bool     `json:"tag"`
		SHA            string   `json:"sha"`
		BeforeSHA      string   `json:"before_sha"`
		Source         string   `json:"source"`
		Status         string   `json:"status"`
		DetailedStatus string   `json:"detailed_status"`
		Stages         []string `json:"stages"`
		CreatedAt      string   `json:"created_at"`
		FinishedAt     string   `json:"finished_at"`
		Duration       int      `json:"duration"`
		QueuedDuration int      `json:"queued_duration"`
		Variables      []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"variables"`
	} `json:"object_attributes"`
	MergeRequest struct {
		ID                 int    `json:"id"`
		IID                int    `json:"iid"`
		Title              string `json:"title"`
		SourceBranch       string `json:"source_branch"`
		SourceProjectID    int    `json:"source_project_id"`
		TargetBranch       string `json:"target_branch"`
		TargetProjectID    int    `json:"target_project_id"`
		State              string `json:"state"`
		MergeRequestStatus string `json:"merge_status"`
		URL                string `json:"url"`
	} `json:"merge_request"`
	User    *EventUser `json:"user"`
	Project struct {
		ID                int             `json:"id"`
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Commit struct {
		ID        string     `json:"id"`
		Message   string     `json:"message"`
		Timestamp *time.Time `json:"timestamp"`
		URL       string     `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commit"`
	Builds []struct {
		ID             int        `json:"id"`
		Stage          string     `json:"stage"`
		Name           string     `json:"name"`
		Status         string     `json:"status"`
		CreatedAt      string     `json:"created_at"`
		StartedAt      string     `json:"started_at"`
		FinishedAt     string     `json:"finished_at"`
		Duration       float64    `json:"duration"`
		QueuedDuration float64    `json:"queued_duration"`
		When           string     `json:"when"`
		Manual         bool       `json:"manual"`
		AllowFailure   bool       `json:"allow_failure"`
		User           *EventUser `json:"user"`
		Runner         struct {
			ID          int      `json:"id"`
			Description string   `json:"description"`
			Active      bool     `json:"active"`
			IsShared    bool     `json:"is_shared"`
			RunnerType  string   `json:"runner_type"`
			Tags        []string `json:"tags"`
		} `json:"runner"`
		ArtifactsFile struct {
			Filename string `json:"filename"`
			Size     int    `json:"size"`
		} `json:"artifacts_file"`
		Environment struct {
			Name           string `json:"name"`
			Action         string `json:"action"`
			DeploymentTier string `json:"deployment_tier"`
		} `json:"environment"`
	} `json:"builds"`
}

// PushEvent represents a push event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#push-events
type PushEvent struct {
	ObjectKind   string `json:"object_kind"`
	Before       string `json:"before"`
	After        string `json:"after"`
	Ref          string `json:"ref"`
	CheckoutSHA  string `json:"checkout_sha"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
	UserEmail    string `json:"user_email"`
	UserAvatar   string `json:"user_avatar"`
	ProjectID    int    `json:"project_id"`
	Project      struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository *Repository `json:"repository"`
	Commits    []*struct {
		ID        string     `json:"id"`
		Message   string     `json:"message"`
		Title     string     `json:"title"`
		Timestamp *time.Time `json:"timestamp"`
		URL       string     `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
}

// ReleaseEvent represents a release event
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#release-events
type ReleaseEvent struct {
	ID          int    `json:"id"`
	CreatedAt   string `json:"created_at"` // Should be *time.Time (see Gitlab issue #21468)
	Description string `json:"description"`
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	ReleasedAt  string `json:"released_at"` // Should be *time.Time (see Gitlab issue #21468)
	ObjectKind  string `json:"object_kind"`
	Project     struct {
		ID                int     `json:"id"`
		Name              string  `json:"name"`
		Description       string  `json:"description"`
		WebURL            string  `json:"web_url"`
		AvatarURL         *string `json:"avatar_url"`
		GitSSHURL         string  `json:"git_ssh_url"`
		GitHTTPURL        string  `json:"git_http_url"`
		Namespace         string  `json:"namespace"`
		VisibilityLevel   int     `json:"visibility_level"`
		PathWithNamespace string  `json:"path_with_namespace"`
		DefaultBranch     string  `json:"default_branch"`
		CIConfigPath      string  `json:"ci_config_path"`
		Homepage          string  `json:"homepage"`
		URL               string  `json:"url"`
		SSHURL            string  `json:"ssh_url"`
		HTTPURL           string  `json:"http_url"`
	} `json:"project"`
	URL    string `json:"url"`
	Action string `json:"action"`
	Assets struct {
		Count int `json:"count"`
		Links []struct {
			ID       int    `json:"id"`
			External bool   `json:"external"`
			LinkType string `json:"link_type"`
			Name     string `json:"name"`
			URL      string `json:"url"`
		} `json:"links"`
		Sources []struct {
			Format string `json:"format"`
			URL    string `json:"url"`
		} `json:"sources"`
	} `json:"assets"`
	Commit struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Title     string `json:"title"`
		Timestamp string `json:"timestamp"` // Should be *time.Time (see Gitlab issue #21468)
		URL       string `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commit"`
}

// SnippetCommentEvent represents a comment on a snippet event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#comment-on-a-code-snippet
type SnippetCommentEvent struct {
	ObjectKind string     `json:"object_kind"`
	User       *EventUser `json:"user"`
	ProjectID  int        `json:"project_id"`
	Project    struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID           int    `json:"id"`
		Note         string `json:"note"`
		NoteableType string `json:"noteable_type"`
		AuthorID     int    `json:"author_id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		ProjectID    int    `json:"project_id"`
		Attachment   string `json:"attachment"`
		LineCode     string `json:"line_code"`
		CommitID     string `json:"commit_id"`
		NoteableID   int    `json:"noteable_id"`
		System       bool   `json:"system"`
		StDiff       *Diff  `json:"st_diff"`
		Description  string `json:"description"`
		URL          string `json:"url"`
	} `json:"object_attributes"`
	Snippet *struct {
		ID                 int    `json:"id"`
		Title              string `json:"title"`
		Content            string `json:"content"`
		AuthorID           int    `json:"author_id"`
		ProjectID          int    `json:"project_id"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
		Filename           string `json:"file_name"`
		ExpiresAt          string `json:"expires_at"`
		Type               string `json:"type"`
		VisibilityLevel    int    `json:"visibility_level"`
		Description        string `json:"description"`
		Secret             bool   `json:"secret"`
		RepositoryReadOnly bool   `json:"repository_read_only"`
	} `json:"snippet"`
}

// SubGroupEvent represents a subgroup event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#subgroup-events
type SubGroupEvent struct {
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	EventName      string     `json:"event_name"`
	Name           string     `json:"name"`
	Path           string     `json:"path"`
	FullPath       string     `json:"full_path"`
	GroupID        int        `json:"group_id"`
	ParentGroupID  int        `json:"parent_group_id"`
	ParentName     string     `json:"parent_name"`
	ParentPath     string     `json:"parent_path"`
	ParentFullPath string     `json:"parent_full_path"`
}

// TagEvent represents a tag event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#tag-events
type TagEvent struct {
	ObjectKind   string `json:"object_kind"`
	Before       string `json:"before"`
	After        string `json:"after"`
	Ref          string `json:"ref"`
	CheckoutSHA  string `json:"checkout_sha"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
	UserAvatar   string `json:"user_avatar"`
	UserEmail    string `json:"user_email"`
	ProjectID    int    `json:"project_id"`
	Message      string `json:"message"`
	Project      struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Repository *Repository `json:"repository"`
	Commits    []*struct {
		ID        string     `json:"id"`
		Message   string     `json:"message"`
		Title     string     `json:"title"`
		Timestamp *time.Time `json:"timestamp"`
		URL       string     `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
}

// WikiPageEvent represents a wiki page event.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/user/project/integrations/webhook_events.html#wiki-page-events
type WikiPageEvent struct {
	ObjectKind string     `json:"object_kind"`
	User       *EventUser `json:"user"`
	Project    struct {
		Name              string          `json:"name"`
		Description       string          `json:"description"`
		AvatarURL         string          `json:"avatar_url"`
		GitSSHURL         string          `json:"git_ssh_url"`
		GitHTTPURL        string          `json:"git_http_url"`
		Namespace         string          `json:"namespace"`
		PathWithNamespace string          `json:"path_with_namespace"`
		DefaultBranch     string          `json:"default_branch"`
		Homepage          string          `json:"homepage"`
		URL               string          `json:"url"`
		SSHURL            string          `json:"ssh_url"`
		HTTPURL           string          `json:"http_url"`
		WebURL            string          `json:"web_url"`
		Visibility        VisibilityValue `json:"visibility"`
	} `json:"project"`
	Wiki struct {
		WebURL            string `json:"web_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		GitHTTPURL        string `json:"git_http_url"`
		PathWithNamespace string `json:"path_with_namespace"`
		DefaultBranch     string `json:"default_branch"`
	} `json:"wiki"`
	ObjectAttributes struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Format  string `json:"format"`
		Message string `json:"message"`
		Slug    string `json:"slug"`
		URL     string `json:"url"`
		Action  string `json:"action"`
	} `json:"object_attributes"`
}

// EventLabel represents a label inside a webhook event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#issue-events
type EventLabel struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Color       string `json:"color"`
	ProjectID   int    `json:"project_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Template    bool   `json:"template"`
	Description string `json:"description"`
	Type        string `json:"type"`
	GroupID     int    `json:"group_id"`
}
