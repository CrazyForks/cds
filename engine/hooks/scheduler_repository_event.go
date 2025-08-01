package hooks

import (
	"context"

	"time"

	"go.opencensus.io/trace"

	"github.com/ovh/cds/sdk"
	cdslog "github.com/ovh/cds/sdk/log"
	"github.com/ovh/cds/sdk/telemetry"
	"github.com/rockbears/log"
)

// Get from queue task execution
func (s *Service) dequeueRepositoryEvent(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			log.ErrorWithStackTrace(ctx, ctx.Err())
			return
		}
		size, err := s.Dao.RepositoryEventQueueLen()
		if err != nil {
			log.Error(ctx, "dequeueRepositoryEvent > Unable to get queueLen: %v", err)
			continue
		}
		log.Debug(ctx, "dequeueRepositoryEvent> current queue size: %d", size)

		if s.Maintenance {
			log.Info(ctx, "Maintenance enable, wait 1 minute. Queue %d", size)
			time.Sleep(1 * time.Minute)
			continue
		}

		// Dequeuing context
		var eventKey string
		if ctx.Err() != nil {
			log.Error(ctx, "%v", err)
			return
		}

		// Get next EventKEY
		if err := s.Cache.DequeueWithContext(ctx, repositoryEventQueue, 250*time.Millisecond, &eventKey); err != nil {
			continue
		}
		s.Dao.dequeuedRepositoryEventIncr()
		if eventKey == "" {
			continue
		}
		log.Info(ctx, "dequeueRepositoryEvent> work on event: %s", eventKey)
		ctx := telemetry.New(ctx, s, "hooks.dequeueRepositoryEvent", nil, trace.SpanKindUnspecified)
		if err := s.manageRepositoryEvent(ctx, eventKey); err != nil {
			log.ErrorWithStackTrace(ctx, err)
			continue
		}

	}
}

func (s *Service) manageRepositoryEvent(ctx context.Context, eventKey string) error {
	ctx, next := telemetry.Span(ctx, "s.manageRepositoryEvent")
	defer next()

	// Load the event
	var hre sdk.HookRepositoryEvent
	find, err := s.Cache.Get(eventKey, &hre)
	if err != nil {
		log.Error(ctx, "dequeueRepositoryEvent> cannot get repository event from cache %s: %v", eventKey, err)
	}
	if !find {
		return nil
	}

	if hre.Initiator == nil && hre.DeprecatedUserID != "" {
		hre.Initiator = &sdk.V2Initiator{UserID: hre.DeprecatedUserID}
	}

	ctx = context.WithValue(ctx, cdslog.HookEventID, hre.UUID)
	ctx = context.WithValue(ctx, cdslog.VCSServer, hre.VCSServerName)
	ctx = context.WithValue(ctx, cdslog.Repository, hre.RepositoryName)

	telemetry.Current(ctx,
		telemetry.Tag(telemetry.TagVCSServer, hre.VCSServerName),
		telemetry.Tag(telemetry.TagRepository, hre.RepositoryName),
		telemetry.Tag(telemetry.TagEventID, hre.UUID))

	b, err := s.Dao.LockRepositoryEvent(hre.VCSServerName, hre.RepositoryName, hre.UUID)
	if err != nil {
		return sdk.WrapError(err, "unable to lock hook event %s", hre.GetFullName())
	}
	defer s.Dao.UnlockRepositoryEvent(hre.VCSServerName, hre.RepositoryName, hre.UUID)

	if !b {
		// reenqueue
		if err := s.Dao.EnqueueRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "unable to reenqueue repository event")
		}
	}

	find, err = s.Cache.Get(eventKey, &hre)
	if err != nil {
		log.Error(ctx, "dequeueRepositoryEvent> cannot get repository event from cache %s: %v", eventKey, err)
	}
	if !find {
		return nil
	}

	// Load the repository
	repoKey := s.Dao.GetRepositoryMemberKey(hre.VCSServerName, hre.RepositoryName)
	repo := s.Dao.FindRepository(ctx, repoKey)
	if repo == nil {
		log.Error(ctx, "dequeueRepositoryEvent failed: Repository %s not found - deleting this event", repoKey)
		hre.LastError = "Internal Error: Repository not found"
		hre.NbErrors++
		hre.Status = sdk.HookEventStatusError
		if err := s.Dao.SaveRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "norepo > unable to save repository event: %s", hre.GetFullName())
		}
		if err := s.Dao.RemoveRepositoryEventFromInProgressList(ctx, hre.UUID); err != nil {
			return sdk.WrapError(err, "norepo > unable to remove event %s from inprogress list", hre.GetFullName())
		}
		return nil
	}

	if repo.Stopped {
		hre.LastError = "Event skipped. Repository hook has been stopped."
		hre.NbErrors++
		hre.Status = sdk.HookEventStatusSkipped
		if err := s.Dao.SaveRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "stopped > unable to save repository event %s", hre.GetFullName())
		}
		if err := s.Dao.RemoveRepositoryEventFromInProgressList(ctx, hre.UUID); err != nil {
			return sdk.WrapError(err, "stopped > unable to remove event %s from inprogress list", hre.GetFullName())
		}
		return nil
	}
	if hre.NbErrors >= s.Cfg.RetryError {
		log.Info(ctx, "dequeueRepositoryEvent> Event %s stopped: to many errors:%d lastError:%s", hre.GetFullName(), hre.NbErrors, hre.LastError)
		hre.Status = sdk.HookEventStatusError
		if err := s.Dao.SaveRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "maxerror > unable to save repository event %s", hre.GetFullName())
		}
		if err := s.Dao.RemoveRepositoryEventFromInProgressList(ctx, hre.UUID); err != nil {
			return sdk.WrapError(err, "maxerror > unable to remove event %s from inprogress list", hre.GetFullName())
		}
		return nil
	}

	if err := s.executeEvent(ctx, &hre); err != nil {
		log.ErrorWithStackTrace(ctx, err)
		log.Warn(ctx, "dequeueRepositoryEvent> %s failed err[%d]: %v", hre.GetFullName(), hre.NbErrors, err)
		hre.LastError = err.Error()
		hre.NbErrors++
		if err := s.Dao.SaveRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "unable to save repository event %s", hre.GetFullName())
		}
		if err := s.Dao.EnqueueRepositoryEvent(ctx, &hre); err != nil {
			return sdk.WrapError(err, "unable to enqueue repository event %s", hre.GetFullName())
		}
	}
	return nil
}

func (s *Service) executeEvent(ctx context.Context, hre *sdk.HookRepositoryEvent) error {
	ctx, next := telemetry.Span(ctx, "s.executeEvent")
	defer next()
	defer func() {
		if hre.EventName != sdk.WorkflowHookEventNameWorkflowRun {
			if err := s.pushInsightReport(ctx, hre); err != nil {
				log.ErrorWithStackTrace(ctx, err)
			}
		}
	}()

	switch hre.Status {
	// Start processing event
	case sdk.HookEventStatusScheduled:
		hre.ProcessingTimestamp = time.Now().UnixNano()
		hre.LastError = ""
		hre.NbErrors = 0
		switch hre.EventName {
		case sdk.WorkflowHookEventNamePush, sdk.WorkflowHookEventNameManual:
			// analyze have to be trigger only on push event
			hre.Status = sdk.HookEventStatusAnalysis
			if err := s.triggerAnalyses(ctx, hre); err != nil {
				return sdk.WrapError(err, "unable to trigger analyses")
			}
		// For PR event, waiting exisiting analysis before searching for hooks
		case sdk.WorkflowHookEventNamePullRequest:
			// Check push analysis to be completed
			hre.Status = sdk.HookEventStatusCheckAnalysis
			if err := s.triggerCheckAnalyses(ctx, hre); err != nil {
				return sdk.WrapError(err, "unable to check analyses")
			}
		default:
			// Retrieve workflow to trigger
			hre.Status = sdk.HookEventStatusWorkflowHooks
			if err := s.triggerGetWorkflowHooks(ctx, hre); err != nil {
				return sdk.WrapError(err, "unable to trigger workflow hooks")
			}
		}
		// Check if all analysis are ended
	case sdk.HookEventStatusAnalysis:
		if err := s.triggerAnalyses(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to trigger analyses")
		}
	case sdk.HookEventStatusCheckAnalysis:
		if err := s.triggerCheckAnalyses(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to check analyses")
		}
		// Retrieve workflow hooks
	case sdk.HookEventStatusWorkflowHooks:
		if err := s.triggerGetWorkflowHooks(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to trigger workflow hooks")
		}
		// Retrieve signing key if we don't have it
	case sdk.HookEventStatusSignKey:
		if err := s.triggerGetSigningKey(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to get signing key")
		}
		// Compute git info ( semver )
	case sdk.HookEventStatusGitInfo:
		if err := s.triggerGetGitInfo(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to get git info")
		}
		// Trigger workflows
	case sdk.HookEventStatusWorkflow:
		if err := s.triggerWorkflows(ctx, hre); err != nil {
			return sdk.WrapError(err, "unable to trigger workflow")
		}
	case sdk.HookEventStatusDone, sdk.HookEventStatusSkipped, sdk.HookEventStatusError:
		// Remove event from inprogressList
		if err := s.Dao.RemoveRepositoryEventFromInProgressList(ctx, hre.UUID); err != nil {
			log.Error(ctx, "executeEvent >unable to remove event %s from inprogress list: %v", hre.UUID, err)
		}
	}
	return nil
}
