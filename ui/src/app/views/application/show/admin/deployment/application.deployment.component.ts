import { ChangeDetectionStrategy, ChangeDetectorRef, Component, Input } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { Store } from '@ngxs/store';
import {
    AddApplicationDeployment,
    DeleteApplicationDeployment,
    UpdateApplicationDeployment
} from 'app/store/applications.action';
import { finalize } from 'rxjs/operators';
import { Application } from 'app/model/application.model';
import { ProjectIntegration } from 'app/model/integration.model';
import { Project } from 'app/model/project.model';
import { ToastService } from 'app/shared/toast/ToastService';

@Component({
    selector: 'app-application-deployment',
    templateUrl: './application.deployment.html',
    styleUrls: ['./application.deployment.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ApplicationDeploymentComponent {
    @Input() project: Project;
    @Input() application: Application;
    @Input() editMode: boolean;
    @Input() deploymentIntegrations: Array<ProjectIntegration>;

    selectedIntegration: ProjectIntegration;
    loadingBtn = false;

    constructor(
        private _toast: ToastService,
        public _translate: TranslateService,
        private store: Store,
        private _cd: ChangeDetectorRef
    ) { }

    getIntegrationNames(): Array<string> {
        if (this.application.deployment_strategies) {
            return Object.keys(this.application.deployment_strategies);
        }
        return null;
    }

    clickDeleteIntegration(pfName: string) {
        this.loadingBtn = true;
        this.store.dispatch(new DeleteApplicationDeployment({
            projectKey: this.project.key,
            applicationName: this.application.name,
            integrationName: pfName
        })).pipe(finalize(() => {
            this.loadingBtn = false;
            this._cd.markForCheck();
        }))
            .subscribe(() => {
                if (this.editMode) {
                    this._toast.info('', this._translate.instant('application_ascode_updated'));
                } else {
                    this._toast.success('', this._translate.instant('application_integration_deleted'));
                }
            });
    }

    updateIntegration(pfName: string) {
        this.loadingBtn = true;
        this.store.dispatch(new UpdateApplicationDeployment({
            projectKey: this.project.key,
            applicationName: this.application.name,
            deploymentName: pfName,
            config: this.application.deployment_strategies[pfName]
        })).pipe(finalize(() => {
            this.loadingBtn = false;
            this._cd.markForCheck();
        }))
            .subscribe(() => {
                if (this.editMode) {
                    this._toast.info('', this._translate.instant('application_ascode_updated'));
                } else {
                    this._toast.success('', this._translate.instant('application_integration_updated'));
                }
            });
    }

    addIntegration() {
        this.loadingBtn = true;
        if (this.selectedIntegration.model) {
            this.store.dispatch(new AddApplicationDeployment({
                projectKey: this.project.key,
                applicationName: this.application.name,
                integration: this.selectedIntegration
            })).pipe(finalize(() => {
                this.loadingBtn = false;
                this._cd.markForCheck();
            }))
                .subscribe(() => {
                    this.selectedIntegration = null;
                    if (this.editMode) {
                        this._toast.info('', this._translate.instant('application_ascode_updated'));
                    } else {
                        this._toast.success('', this._translate.instant('application_integration_added'));
                    }
                });
        }
    }
}
