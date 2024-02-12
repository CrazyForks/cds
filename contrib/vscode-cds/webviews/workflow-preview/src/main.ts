import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';
import { PreviewModule } from './app/preview.module';

platformBrowserDynamic().bootstrapModule(PreviewModule)
  .catch(err => console.error(err));