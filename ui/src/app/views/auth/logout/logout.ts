import { ChangeDetectionStrategy, Component } from '@angular/core';

@Component({
    selector: 'app-auth-logout',
    templateUrl: './logout.html',
    styleUrls: ['./logout.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LogoutComponent { }
