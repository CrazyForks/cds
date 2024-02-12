import { Component, HostListener } from "@angular/core";

@Component({
    selector: 'app-root',
    templateUrl: './preview.component.html',
    styleUrls: ['./preview.component.scss']
  })
  export class PreviewComponent {
    
    fileValue: String;
  
    constructor() {
      this.fileValue = '';
    }
  
    @HostListener('window:message', ['$event'])
    onRefresh(e: MessageEvent) {
      console.log(e);
      if (e.data.type === 'refresh') {
        this.fileValue = e.data.value;
      }
    }
  }