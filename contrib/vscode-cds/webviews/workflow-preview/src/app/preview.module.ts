import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { PreviewComponent } from './preview.component';
import { WorkflowGraphModule} from 'workflow-graph';

@NgModule({
  declarations: [
    PreviewComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    WorkflowGraphModule
  ],
  providers: [],
  bootstrap: [PreviewComponent]
})
export class PreviewModule { }