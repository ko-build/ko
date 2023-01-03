# go-gitlab

A GitLab API client enabling Go programs to interact with GitLab in a simple and uniform way

[![Build Status](https://github.com/xanzy/go-gitlab/workflows/Lint%20and%20Test/badge.svg)](https://github.com/xanzy/go-gitlab/actions?workflow=Lint%20and%20Test)
[![Sourcegraph](https://sourcegraph.com/github.com/xanzy/go-gitlab/-/badge.svg)](https://sourcegraph.com/github.com/xanzy/go-gitlab?badge)
[![GoDoc](https://godoc.org/github.com/xanzy/go-gitlab?status.svg)](https://godoc.org/github.com/xanzy/go-gitlab)
[![Go Report Card](https://goreportcard.com/badge/github.com/xanzy/go-gitlab)](https://goreportcard.com/report/github.com/xanzy/go-gitlab)

## NOTE

Release v0.6.0 (released on 25-08-2017) no longer supports the older V3 GitLab API. If
you need V3 support, please use the `f-api-v3` branch. This release contains some backwards
incompatible changes that were needed to fully support the V4 GitLab API.

## Coverage

This API client package covers most of the existing GitLab API calls and is updated regularly
to add new and/or missing endpoints. Currently, the following services are supported:

- [x] Applications
- [x] Award Emojis
- [x] Branches
- [x] Broadcast Messages
- [x] Commits
- [x] Container Registry
- [x] Custom Attributes
- [x] Deploy Keys
- [x] Deployments
- [x] Discussions (threaded comments)
- [x] Environments
- [x] Epic Issues
- [x] Epics
- [x] Error Tracking
- [x] Events
- [x] Feature Flags
- [x] Geo Nodes
- [x] Generic Packages
- [x] GitLab CI Config Templates
- [x] Gitignores Templates
- [x] Group Access Requests
- [x] Group Issue Boards
- [x] Group Members
- [x] Group Milestones
- [x] Group Wikis
- [x] Group-Level Variables
- [x] Groups
- [x] Instance Clusters
- [x] Invites
- [x] Issue Boards
- [x] Issues
- [x] Jobs
- [x] Keys
- [x] Labels
- [x] License
- [x] Markdown
- [x] Merge Request Approvals
- [x] Merge Requests
- [x] Namespaces
- [x] Notes (comments)
- [x] Notification Settings
- [x] Open Source License Templates
- [x] Packages
- [x] Pages
- [x] Pages Domains
- [x] Personal Access Tokens
- [x] Pipeline Schedules
- [x] Pipeline Triggers
- [x] Pipelines
- [x] Plan limits
- [x] Project Access Requests
- [x] Project Badges
- [x] Project Clusters
- [x] Project Import/export
- [x] Project Members
- [x] Project Milestones
- [x] Project Snippets
- [x] Project Vulnerabilities
- [x] Project-Level Variables
- [x] Projects (including setting Webhooks)
- [x] Protected Branches
- [x] Protected Environments
- [x] Protected Tags
- [x] Repositories
- [x] Repository Files
- [x] Repository Submodules
- [x] Runners
- [x] Search
- [x] Services
- [x] Settings
- [x] Sidekiq Metrics
- [x] System Hooks
- [x] Tags
- [x] Todos
- [x] Topics
- [x] Users
- [x] Validate CI Configuration
- [x] Version
- [x] Wikis

## Usage

```go
import "github.com/xanzy/go-gitlab"
```

Construct a new GitLab client, then use the various services on the client to
access different parts of the GitLab API. For example, to list all
users:

```go
git, err := gitlab.NewClient("yourtokengoeshere")
if err != nil {
  log.Fatalf("Failed to create client: %v", err)
}
users, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{})
```

There are a few `With...` option functions that can be used to customize
the API client. For example, to set a custom base URL:

```go
git, err := gitlab.NewClient("yourtokengoeshere", gitlab.WithBaseURL("https://git.mydomain.com/api/v4"))
if err != nil {
  log.Fatalf("Failed to create client: %v", err)
}
users, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{})
```

Some API methods have optional parameters that can be passed. For example,
to list all projects for user "svanharmelen":

```go
git := gitlab.NewClient("yourtokengoeshere")
opt := &gitlab.ListProjectsOptions{Search: gitlab.String("svanharmelen")}
projects, _, err := git.Projects.ListProjects(opt)
```

### Examples

The [examples](https://github.com/xanzy/go-gitlab/tree/master/examples) directory
contains a couple for clear examples, of which one is partially listed here as well:

```go
package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func main() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String("My Project"),
		Description:          gitlab.String("Just a test project to play with"),
		MergeRequestsEnabled: gitlab.Bool(true),
		SnippetsEnabled:      gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.PublicVisibility),
	}
	project, _, err := git.Projects.CreateProject(p)
	if err != nil {
		log.Fatal(err)
	}

	// Add a new snippet
	s := &gitlab.CreateProjectSnippetOptions{
		Title:           gitlab.String("Dummy Snippet"),
		FileName:        gitlab.String("snippet.go"),
		Content:         gitlab.String("package main...."),
		Visibility:      gitlab.Visibility(gitlab.PublicVisibility),
	}
	_, _, err = git.ProjectSnippets.CreateSnippet(project.ID, s)
	if err != nil {
		log.Fatal(err)
	}
}
```

For complete usage of go-gitlab, see the full [package docs](https://godoc.org/github.com/xanzy/go-gitlab).

## ToDo

- The biggest thing this package still needs is tests :disappointed:

## Issues

- If you have an issue: report it on the [issue tracker](https://github.com/xanzy/go-gitlab/issues)

## Author

Sander van Harmelen (<sander@vanharmelen.nl>)

## License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>
