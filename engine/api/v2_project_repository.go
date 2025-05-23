package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/ovh/cds/engine/api/event_v2"
	"github.com/ovh/cds/engine/api/project"
	"github.com/ovh/cds/engine/api/repositoriesmanager"
	"github.com/ovh/cds/engine/api/repository"
	"github.com/ovh/cds/engine/api/services"
	"github.com/ovh/cds/engine/gorpmapper"
	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/telemetry"
)

func (api *API) getRepositoryByIdentifier(ctx context.Context, vcsID string, repositoryIdentifier string, opts ...gorpmapper.GetOptionFunc) (*sdk.ProjectRepository, error) {
	ctx, next := telemetry.Span(ctx, "api.getRepositoryByIdentifier")
	defer next()
	var repo *sdk.ProjectRepository
	var err error
	if sdk.IsValidUUID(repositoryIdentifier) {
		repo, err = repository.LoadRepositoryByVCSAndID(ctx, api.mustDB(), vcsID, repositoryIdentifier, opts...)
	} else {
		repo, err = repository.LoadRepositoryByName(ctx, api.mustDB(), vcsID, repositoryIdentifier, opts...)
	}
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (api *API) getProjectRepositoryEventHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
			if err != nil {
				return err
			}

			eventID := vars["eventID"]

			srvs, err := services.LoadAllByType(ctx, api.mustDB(), sdk.TypeHooks)
			if err != nil {
				return err
			}
			if len(srvs) < 1 {
				return sdk.NewErrorFrom(sdk.ErrNotFound, "unable to find hook uservice")
			}
			path := fmt.Sprintf("/v2/repository/event/%s/%s/%s", vcsProject.Name, url.PathEscape(repo.Name), eventID)
			var repositoryEvent sdk.HookRepositoryEvent
			_, code, errHooks := services.NewClient(srvs).DoJSONRequest(ctx, http.MethodGet, path, nil, &repositoryEvent)
			if (errHooks != nil || code >= 400) && code != 404 {
				return sdk.WrapError(errHooks, "unable to get repository event [HTTP: %d]", code)
			}
			return service.WriteJSON(w, repositoryEvent, http.StatusOK)
		}
}

func (api *API) getProjectRepositoryEventsHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
			if err != nil {
				return err
			}

			// Remove hooks
			srvs, err := services.LoadAllByType(ctx, api.mustDB(), sdk.TypeHooks)
			if err != nil {
				return err
			}
			if len(srvs) < 1 {
				return sdk.NewErrorFrom(sdk.ErrNotFound, "unable to find hook uservice")
			}
			path := fmt.Sprintf("/v2/repository/event/%s/%s", vcsProject.Name, url.PathEscape(repo.Name))
			var repositoryEvents []sdk.HookRepositoryEvent
			_, code, errHooks := services.NewClient(srvs).DoJSONRequest(ctx, http.MethodGet, path, nil, &repositoryEvents)
			if (errHooks != nil || code >= 400) && code != 404 {
				return sdk.WrapError(errHooks, "unable to list repository event [HTTP: %d]", code)
			}
			return service.WriteJSON(w, repositoryEvents, http.StatusOK)
		}
}

func (api *API) getProjectRepositoryHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
			if err != nil {
				return err
			}

			return service.WriteJSON(w, repo, http.StatusOK)
		}
}

// deleteProjectRepositoryHandler Delete a repository from a project
func (api *API) deleteProjectRepositoryHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectManage),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
			if err != nil {
				return err
			}

			u := getUserConsumer(ctx)
			if u == nil {
				return sdk.WithStack(sdk.ErrForbidden)
			}

			tx, err := api.mustDB().Begin()
			if err != nil {
				return sdk.WithStack(err)
			}
			defer tx.Rollback() // nolint

			if err := repository.Delete(tx, repo.VCSProjectID, repo.Name); err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return sdk.WithStack(err)
			}

			event_v2.PublishRepositoryEvent(ctx, api.Cache, sdk.EventRepositoryDeleted, pKey, vcsProject.Name, *repo, *u.AuthConsumerUser.AuthentifiedUser)
			return service.WriteMarshal(w, req, vcsProject, http.StatusOK)
		}
}

// postProjectRepositoryHandler Attach a new repository to the given project
func (api *API) postProjectRepositoryHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectManage),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]

			u := getUserConsumer(ctx)
			if u == nil {
				return sdk.WithStack(sdk.ErrForbidden)
			}

			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			vcsProjectWithSecret, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier, gorpmapper.GetOptions.WithDecryption)
			if err != nil {
				return err
			}

			proj, err := project.Load(ctx, api.mustDB(), pKey)
			if err != nil {
				return err
			}

			var repoBody sdk.ProjectRepository
			if err := service.UnmarshalRequest(ctx, req, &repoBody); err != nil {
				return err
			}
			repoBody.ProjectKey = pKey

			tx, err := api.mustDB().Begin()
			if err != nil {
				return sdk.WithStack(err)
			}
			defer tx.Rollback() // nolint

			repoDB := repoBody
			repoDB.VCSProjectID = vcsProjectWithSecret.ID
			repoDB.CreatedBy = getUserConsumer(ctx).GetUsername()
			repoDB.Name = strings.ToLower(repoDB.Name)
			// Insert Repository
			if err := repository.Insert(ctx, tx, &repoDB); err != nil {
				return err
			}

			// Check if repo exist
			vcsClient, err := repositoriesmanager.AuthorizedClient(ctx, tx, api.Cache, pKey, vcsProjectWithSecret.Name)
			if err != nil {
				return err
			}
			vcsRepo, err := vcsClient.RepoByFullname(ctx, repoDB.Name)
			if err != nil {
				return err
			}
			defaultBranch, err := vcsClient.Branch(ctx, repoDB.Name, sdk.VCSBranchFilters{Default: true})
			if err != nil {
				return err
			}

			if vcsProjectWithSecret.Auth.SSHKeyName != "" {
				if vcsRepo.SSHCloneURL == "" {
					return sdk.NewErrorFrom(sdk.ErrInvalidData, "this repo cannot be cloned using ssh.")
				}
				repoDB.CloneURL = vcsRepo.SSHCloneURL
			} else {
				if vcsRepo.HTTPCloneURL == "" {
					return sdk.NewErrorFrom(sdk.ErrInvalidData, "this repo cannot be cloned using https. Please provide a sshkey.")
				}
				repoDB.CloneURL = vcsRepo.HTTPCloneURL
			}
			if err := repository.Update(ctx, tx, &repoDB); err != nil {
				return err
			}

			createAnalysis := createAnalysisRequest{
				proj:          *proj,
				vcsProject:    *vcsProjectWithSecret,
				repo:          repoDB,
				ref:           defaultBranch.ID,
				commit:        defaultBranch.LatestCommit,
				hookEventUUID: "",
				initiator: &sdk.V2Initiator{
					UserID:         u.AuthConsumerUser.AuthentifiedUser.ID,
					User:           u.AuthConsumerUser.AuthentifiedUser.Initiator(),
					IsAdminWithMFA: isAdmin(ctx),
				},
			}
			a, err := api.createAnalyze(ctx, tx, createAnalysis)
			if err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return sdk.WithStack(err)
			}

			// Clean events
			srvs, err := services.LoadAllByType(ctx, api.mustDB(), sdk.TypeHooks)
			if err != nil {
				return err
			}
			if len(srvs) < 1 {
				return sdk.NewErrorFrom(sdk.ErrNotFound, "unable to find hook uservice")
			}
			path := fmt.Sprintf("/v2/repository/event/%s/%s", vcsProjectWithSecret.Name, url.PathEscape(repoDB.Name))
			_, code, errHooks := services.NewClient(srvs).DoJSONRequest(ctx, http.MethodDelete, path, nil, nil)
			if (errHooks != nil || code >= 400) && code != 404 {
				return sdk.WrapError(errHooks, "unable to delete repository events [HTTP: %d]", code)
			}

			event_v2.PublishRepositoryEvent(ctx, api.Cache, sdk.EventRepositoryCreated, pKey, vcsProjectWithSecret.Name, repoDB, *u.AuthConsumerUser.AuthentifiedUser)

			event_v2.PublishAnalysisStart(ctx, api.Cache, vcsProjectWithSecret.Name, repoDB.Name, a)
			return service.WriteMarshal(w, req, repoDB, http.StatusCreated)
		}
}

// getVCSProjectRepositoryAllHandler returns the list of repositories linked to the given vcs/project
func (api *API) getVCSProjectRepositoryAllHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			vars := mux.Vars(r)
			pKey := vars["projectKey"]

			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			repositories, err := repository.LoadAllRepositoriesByVCSProjectID(ctx, api.mustDB(), vcsProject.ID)
			if err != nil {
				return err
			}
			return service.WriteJSON(w, repositories, http.StatusOK)
		}
}

func (api *API) getProjectRepositoryBranchesHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}
			limitS := QueryString(req, "limit")
			limit, err := strconv.Atoi(limitS)
			if limit == 0 || err != nil {
				limit = 50
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			var repoName string
			if sdk.IsValidUUID(repositoryIdentifier) {
				repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
				if err != nil {
					return err
				}
				repoName = repo.Name
			} else {
				repoName = repositoryIdentifier
			}

			client, err := repositoriesmanager.AuthorizedClient(ctx, api.mustDB(), api.Cache, pKey, vcsProject.Name)
			if err != nil {
				return err
			}

			branches, err := client.Branches(ctx, repoName, sdk.VCSBranchesFilter{Limit: int64(limit)})
			if err != nil {
				return err
			}

			return service.WriteJSON(w, branches, http.StatusOK)
		}
}

func (api *API) getProjectRepositoryTagsHandler() ([]service.RbacChecker, service.Handler) {
	return service.RBAC(api.projectRead),
		func(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
			vars := mux.Vars(req)
			pKey := vars["projectKey"]
			vcsIdentifier, err := url.PathUnescape(vars["vcsIdentifier"])
			if err != nil {
				return sdk.NewError(sdk.ErrWrongRequest, err)
			}
			repositoryIdentifier, err := url.PathUnescape(vars["repositoryIdentifier"])
			if err != nil {
				return sdk.WithStack(err)
			}

			vcsProject, err := api.getVCSByIdentifier(ctx, pKey, vcsIdentifier)
			if err != nil {
				return err
			}

			var repoName string
			if sdk.IsValidUUID(repositoryIdentifier) {
				repo, err := api.getRepositoryByIdentifier(ctx, vcsProject.ID, repositoryIdentifier)
				if err != nil {
					return err
				}
				repoName = repo.Name
			} else {
				repoName = repositoryIdentifier
			}

			client, err := repositoriesmanager.AuthorizedClient(ctx, api.mustDB(), api.Cache, pKey, vcsProject.Name)
			if err != nil {
				return err
			}

			tags, err := client.Tags(ctx, repoName)
			if err != nil {
				return err
			}

			return service.WriteJSON(w, tags, http.StatusOK)
		}
}
