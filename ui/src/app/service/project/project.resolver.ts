import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Store } from '@ngxs/store';
import { LoadOpts } from 'app/model/project.model';
import { RouterService } from 'app/service/router/router.service';
import { FetchProject } from 'app/store/project.action';
import { ProjectState } from 'app/store/project.state';
import { Observable } from 'rxjs';
import { switchMap } from 'rxjs/operators';

@Injectable()
export class ProjectResolver {

    constructor(
        private store: Store,
        private routerService: RouterService
    ) { }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<any> | Promise<any> | any {
        let params = this.routerService.getRouteSnapshotParams({}, state.root);
        let opts = [
            new LoadOpts('withApplicationNames', 'application_names'),
            new LoadOpts('withPipelineNames', 'pipeline_names'),
            new LoadOpts('withWorkflowNames', 'workflow_names'),
            new LoadOpts('withEnvironmentNames', 'environment_names'),
            new LoadOpts('withLabels', 'labels')
        ];

        return this.store.dispatch(new FetchProject({
            projectKey: params['key'],
            opts
        })).pipe(
            switchMap(() => this.store.selectOnce(ProjectState.projectSnapshot)),
        );
    }
}

@Injectable()
export class ProjectForWorkflowResolver {

    constructor(private store: Store, private routerService: RouterService) { }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<any> | Promise<any> | any {
        let params = this.routerService.getRouteSnapshotParams({}, state.root);

        let opts = [
            new LoadOpts('withWorkflowNames', 'workflow_names'),
            new LoadOpts('withPipelineNames', 'pipeline_names'),
            new LoadOpts('withApplicationNames', 'application_names'),
            new LoadOpts('withEnvironmentNames', 'environment_names'),
            new LoadOpts('withEnvironments', 'environments'),
            new LoadOpts('withLabels', 'labels'),
            new LoadOpts('withKeys', 'keys'),
            new LoadOpts('withIntegrations', 'integrations')
        ];

        return this.store.dispatch(new FetchProject({
            projectKey: params['key'],
            opts
        })).pipe(
            switchMap(() => this.store.selectOnce(ProjectState.projectSnapshot))
        );
    }
}

@Injectable()
export class ProjectForApplicationResolver {

    constructor(private store: Store, private routerService: RouterService) { }

    resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<any> | Promise<any> | any {
        let params = this.routerService.getRouteSnapshotParams({}, state.root);
        let opts = [
            new LoadOpts('withWorkflowNames', 'workflow_names'),
            new LoadOpts('withPipelineNames', 'pipeline_names'),
            new LoadOpts('withApplicationNames', 'application_names'),
            new LoadOpts('withEnvironmentNames', 'environment_names')
        ];

        return this.store.dispatch(new FetchProject({
            projectKey: params['key'],
            opts
        })).pipe(
            switchMap(() => this.store.selectOnce(ProjectState.projectSnapshot))
        );
    }
}
