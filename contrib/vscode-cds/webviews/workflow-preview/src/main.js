"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const platform_browser_dynamic_1 = require("@angular/platform-browser-dynamic");
const preview_module_1 = require("./app/preview.module");
(0, platform_browser_dynamic_1.platformBrowserDynamic)().bootstrapModule(preview_module_1.PreviewModule)
    .catch(err => console.error(err));
//# sourceMappingURL=main.js.map