(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[139],{73570:function(t,e,n){var s=n(49237),r="  ";function a(t){var e=typeof t;return t instanceof Array?"array":"string"==e?"string":"boolean"==e?"boolean":"number"==e?"number":"undefined"==e||null===t?"null":"hash"}function i(t,e){var n=a(t);switch(n){case"array":o(t,e);break;case"hash":l(t,e);break;case"string":u(t,e);break;case"null":e.push("null");break;case"number":e.push(t.toString());break;case"boolean":e.push(t?"true":"false");break}}function o(t,e){for(var n=0;n<t.length;n++){var s=t[n],a=[];i(s,a);for(var o=0;o<a.length;o++)e.push((0==o?"- ":r)+a[o])}}function l(t,e){for(var n in t){var s=[];if(t.hasOwnProperty(n)){var o=t[n];i(o,s);var l=a(o);if("string"==l||"null"==l||"number"==l||"boolean"==l)e.push(c(n)+": "+s[0]);else{e.push(c(n)+": ");for(var u=0;u<s.length;u++)e.push(r+s[u])}}}}function c(t){return t.match(/^[\w]+$/)?t:s.requiresDoubleQuoting(t)?s.escapeWithDoubleQuotes(t):s.requiresSingleQuoting(t)?s.escapeWithSingleQuotes(t):t}function u(t,e){e.push(c(t))}var d=function(t){"string"==typeof t&&(t=JSON.parse(t));var e=[];return i(t,e),e.join("\n")};t.exports=d},22330:function(t,e,n){"use strict";n.d(e,{Z:function(){return d}});var s=function(){var t=this,e=t._self._c;return e("div",{staticClass:"code-view"},[t.isReady?e("div",{staticClass:"code-view-content"},[t.isLoading||t.isEmpty?t._e():e("KCard",{attrs:{title:t.title,"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("Prism",{staticClass:"code-block",attrs:{language:t.lang,code:t.codeContent}})]},proxy:!0},{key:"actions",fn:function(){return[t.content?e("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function({copyToClipboard:n}){return[e("KPop",{attrs:{placement:"bottom"},scopedSlots:t._u([{key:"content",fn:function(){return[e("div",[e("p",[t._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[e("KButton",{attrs:{appearance:"primary"},on:{click:()=>{n(t.codeContent)}}},[t._v(" "+t._s(t.copyButtonText)+" ")])],1)]}}],null,!1,4222532487)}):t._e()]},proxy:!0}],null,!1,1362931160)})],1):t._e(),!0===t.loaders?e("div",[t.isLoading?e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):t._e(),t.isEmpty&&!t.isLoading?e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" There is no data to display. ")]},proxy:!0}],null,!1,1612658095)}):t._e(),t.hasError?e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1):t._e()])},r=[],a=n(90236),i=n.n(a),o={name:"CodeView",components:{Prism:i()},props:{lang:{type:String,required:!0},copyButtonText:{type:String,default:"Copy to Clipboard"},title:{type:String,default:null},content:{type:String,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},computed:{isReady(){return!this.isEmpty&&!this.hasError&&!this.isLoading},codeContent(){const t=this.content;return t}}},l=o,c=n(1001),u=(0,c.Z)(l,s,r,!1,null,"575cf2bc",null),d=u.exports},7001:function(t,e,n){"use strict";n.d(e,{Z:function(){return h}});var s=function(){var t=this,e=t._self._c;return e("div",{staticClass:"tab-container",attrs:{"data-testid":"tab-container"}},[t.$slots.tabHeader&&t.isReady?e("header",{staticClass:"tab__header"},[t._t("tabHeader")],2):t._e(),e("div",{staticClass:"tab__content-container"},[t.isReady?e("KTabs",{attrs:{tabs:t.tabs},on:{changed:e=>t.switchTab(e)},scopedSlots:t._u([t._l(t.tabsSlots,(function(e){return{key:e,fn:function(){return[t._t(e)]},proxy:!0}})),{key:"warnings-anchor",fn:function(){return[e("span",{staticClass:"flex items-center with-warnings"},[e("KIcon",{staticClass:"mr-1",attrs:{color:"var(--yellow-400)",icon:"warning","secondary-color":"var(--black-75)",size:"16"}}),e("span",[t._v(" Warnings ")])],1)]},proxy:!0}],null,!0),model:{value:t.tabState,callback:function(e){t.tabState=e},expression:"tabState"}}):t._e(),!0===t.loaders?e("div",[t.isLoading?e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):t._e(),t.hasError?e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1):t._e()],1)])},r=[],a=n(30037),i=n.n(a),o=n(89340),l=n(70878),c={name:"TabsWidget",components:{KEmptyState:i()},props:{loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},tabs:{type:Array,required:!0},hasBorder:{type:Boolean,default:!1},initialTabOverride:{type:String,default:null}},data(){return{tabState:this.initialTabOverride&&`#${this.initialTabOverride}`}},computed:{tabsSlots(){return this.tabs.map((t=>t.hash.replace("#","")))},isReady(){return!1===this.loaders||!this.isEmpty&&!this.hasError&&!this.isLoading}},methods:{switchTab(t){o.fy.logger.info(l.T.TABS_TAB_CHANGE,{data:{newTab:t}}),this.$emit("onTabChange",t)}}},u=c,d=n(1001),p=(0,d.Z)(u,s,r,!1,null,"cbdcede8",null),h=p.exports},69328:function(t,e,n){"use strict";n.d(e,{Z:function(){return f}});var s=function(){var t=this,e=t._self._c;return t.shouldStart?e("div",{staticClass:"scanner"},[e("div",{staticClass:"scanner-content"},[e("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[t.isRunning?e("div",{staticClass:"card-icon mb-3"},[e("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1):t._e(),t.isComplete&&!1===t.hasError&&!1===t.isRunning?e("div",{staticClass:"card-icon mb-3"},[e("IconSuccess")],1):t._e(),t.isRunning?t._t("loading-title"):t._e(),!1===t.isRunning?e("div",[t.hasError?t._t("error-title"):t._e(),t.isComplete&&!1===t.hasError?t._t("complete-title"):t._e()],2):t._e()]},proxy:!0},{key:"message",fn:function(){return[t.isRunning?t._t("loading-content"):t._e(),!1===t.isRunning?e("div",[t.hasError?t._t("error-content"):t._e(),t.isComplete&&!1===t.hasError?t._t("complete-content"):t._e()],2):t._e()]},proxy:!0}],null,!0)})],1)]):t._e()},r=[],a=function(){var t=this,e=t._self._c;return e("i",{staticClass:"card-icon icon-success mb-3",attrs:{role:"img"}},[t._v(" ✓ ")])},i=[],o={},l=o,c=n(1001),u=(0,c.Z)(l,a,i,!1,null,"38718532",null),d=u.exports,p={name:"EntityScanner",components:{IconSuccess:d},props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},shouldStart:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},data(){return{i:0,isRunning:!1,isComplete:!1,intervalId:null}},watch:{shouldStart(t,e){t!==e&&!0===t&&this.runScanner()}},mounted(){!0===this.shouldStart&&this.runScanner()},beforeDestroy(){clearInterval(this.intervalId)},methods:{runScanner(){this.isRunning=!0,this.isComplete=!1,this.intervalId=setInterval((()=>{this.i++,this.loaderFunction(),this.i!==this.retries&&!0!==this.canComplete||(clearInterval(this.intervalId),this.isRunning=!1,this.isComplete=!0,this.$emit("hideSiblings",!0))}),this.interval)}}},h=p,_=(0,c.Z)(h,s,r,!1,null,"4754238e",null),f=_.exports},5035:function(t,e,n){"use strict";n.d(e,{Z:function(){return c}});var s=function(){var t=this,e=t._self._c;return e("div",{staticClass:"form-line-wrapper"},[e("div",{staticClass:"form-line",class:{"has-equal-cols":t.equalCols}},[t.hideLabelCol?t._e():e("div",{staticClass:"form-line__col"},[e("label",{staticClass:"k-input-label",attrs:{for:t.forAttr}},[t._v(" "+t._s(t.title)+": ")])]),e("div",{staticClass:"form-line__col",class:{"is-inline":t.allInline,"is-shifted-right":t.shiftRight}},[t._t("default")],2)])])},r=[],a={name:"FormFragment",props:{title:{type:String,required:!1,default:null},forAttr:{type:String,required:!1,default:null},allInline:{type:Boolean,default:!1},hideLabelCol:{type:Boolean,default:!1},equalCols:{type:Boolean,default:!1},shiftRight:{type:Boolean,default:!1}}},i=a,o=n(1001),l=(0,o.Z)(i,s,r,!1,null,"757dc37d",null),c=l.exports},5872:function(t,e,n){"use strict";n.d(e,{Z:function(){return d}});var s=function(){var t=this,e=t._self._c;return e("div",{staticClass:"wizard-steps"},[e("div",{staticClass:"wizard-steps__content-wrapper"},[e("header",{staticClass:"wizard-steps__indicator"},[e("ul",{staticClass:"wizard-steps__indicator__controls",attrs:{role:"tablist","aria-label":"steptabs"}},t._l(t.steps,(function(n,s){return e("li",{key:n.slug,staticClass:"wizard-steps__indicator__item",class:{"is-complete":s<=t.start},attrs:{"aria-selected":t.step===n.slug?"true":"false","aria-controls":`wizard-steps__content__item--${s}`}},[e("span",[t._v(" "+t._s(n.label)+" ")])])})),0)]),e("div",{staticClass:"wizard-steps__content"},[e("form",{ref:"wizardForm",attrs:{autocomplete:"off"}},t._l(t.steps,(function(n,s){return e("div",{key:n.slug,staticClass:"wizard-steps__content__item",attrs:{id:`wizard-steps__content__item--${s}`,"aria-labelledby":`wizard-steps__content__item--${s}`,role:"tabpanel",tabindex:"0"}},[t.step===n.slug?t._t(n.slug):t._e()],2)})),0)]),t.footerEnabled?e("footer",{staticClass:"wizard-steps__footer"},[e("KButton",{directives:[{name:"show",rawName:"v-show",value:!t.indexCanReverse,expression:"!indexCanReverse"}],attrs:{appearance:"secondary"},on:{click:t.goToPrevStep}},[t._v(" ‹ Previous ")]),e("KButton",{directives:[{name:"show",rawName:"v-show",value:!t.indexCanAdvance,expression:"!indexCanAdvance"}],attrs:{disabled:t.nextDisabled,appearance:"primary"},on:{click:t.goToNextStep}},[t._v(" Next › ")])],1):t._e()]),e("aside",{staticClass:"wizard-steps__sidebar"},[e("div",{staticClass:"wizard-steps__sidebar__content"},t._l(t.sidebarContent,(function(n,s){return e("div",{key:n.name,staticClass:"wizard-steps__sidebar__item",class:`wizard-steps__sidebar__item--${s}`},[t._t(n.name)],2)})),0)])])},r=[],a=n(70538),i=a.ZP.extend({methods:{updateQuery(t,e){const n=this.$router,s=this.$route;s.query?n.push({query:Object.assign({},s.query,{[t]:e})}).catch((()=>{})):n.push({query:{[t]:e}}).catch((()=>{}))}}}),o={mixins:[i],props:{steps:{type:Array,default:()=>{}},sidebarContent:{type:Array,required:!0,default:()=>{}},footerEnabled:{type:Boolean,default:!0},nextDisabled:{type:Boolean,default:!0}},data(){return{start:0}},computed:{step:{get(){return this.steps[this.start].slug},set(t){return this.steps[t].slug}},indexCanAdvance(){return this.start>=this.steps.length-1},indexCanReverse(){return this.start<=0}},watch:{"$route.query.step"(t=0){this.start!==t&&(this.start=t,this.$emit("goToNextStep",t))}},mounted(){this.resetProcess(),this.setStartingStep()},methods:{goToStep(t){this.start=t,this.updateQuery("step",t),this.$emit("goToStep",this.step)},goToNextStep(){this.start++,this.updateQuery("step",this.start),this.$emit("goToNextStep",this.step)},goToPrevStep(){this.start--,this.updateQuery("step",this.start),this.$emit("goToPrevStep",this.step)},setStartingStep(){const t=this.$route.query.step;this.start=t||0},resetProcess(){this.start=0,this.goToStep(0),localStorage.removeItem("storedFormData");const t=this.$refs.wizardForm.querySelectorAll('input[type="text"]');t.forEach((t=>{t.setAttribute("value","")}))}}},l=o,c=n(1001),u=(0,c.Z)(l,s,r,!1,null,"0ffab7b9",null),d=u.exports},70878:function(t,e,n){"use strict";n.d(e,{T:function(){return s}});const s={PAGINATION_PREVIOUS_BUTTON_CLICKED:"pagination-previous-button-clicked",PAGINATION_NEXT_BUTTON_CLICKED:"pagination-next-button-clicked",SIDEBAR_ITEM_CLICKED:"sidebar-item-clicked",TABLE_REFRESH_BUTTON_CLICKED:"table-refresh-button-clicked",TABS_TAB_CHANGE:"tabs-tab-change",CREATE_MESH_CLICKED:"create-mesh-clicked",CREATE_DATA_PLANE_PROXY_CLICKED:"create-data-plane-proxy-clicked"}},49237:function(t,e,n){var s,r;r=n(11665),s=function(){var t;function e(){}return e.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(t=String.fromCharCode)(133),t(160),t(8232),t(8233)],e.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],e.MAPPING_ESCAPEES_TO_ESCAPED=function(){var t,n,s,r;for(s={},t=n=0,r=e.LIST_ESCAPEES.length;0<=r?n<r:n>r;t=0<=r?++n:--n)s[e.LIST_ESCAPEES[t]]=e.LIST_ESCAPED[t];return s}(),e.PATTERN_CHARACTERS_TO_ESCAPE=new r("[\\x00-\\x1f]|Â|Â |â¨|â©"),e.PATTERN_MAPPING_ESCAPEES=new r(e.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),e.PATTERN_SINGLE_QUOTING=new r("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),e.requiresDoubleQuoting=function(t){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(t)},e.escapeWithDoubleQuotes=function(t){var e;return e=this.PATTERN_MAPPING_ESCAPEES.replace(t,function(t){return function(e){return t.MAPPING_ESCAPEES_TO_ESCAPED[e]}}(this)),'"'+e+'"'},e.requiresSingleQuoting=function(t){return this.PATTERN_SINGLE_QUOTING.test(t)},e.escapeWithSingleQuotes=function(t){return"'"+t.replace(/'/g,"''")+"'"},e}(),t.exports=s},11665:function(t){var e;e=function(){function t(t,e){var n,s,r,a,i,o,l,c,u;null==e&&(e=""),r="",i=t.length,o=null,s=0,a=0;while(a<i){if(n=t.charAt(a),"\\"===n)r+=t.slice(a,+(a+1)+1||9e9),a++;else if("("===n)if(a<i-2)if(c=t.slice(a,+(a+2)+1||9e9),"(?:"===c)a+=2,r+=c;else if("(?<"===c){s++,a+=2,l="";while(a+1<i){if(u=t.charAt(a+1),">"===u){r+="(",a++,l.length>0&&(null==o&&(o={}),o[l]=s);break}l+=u,a++}}else r+=n,s++;else r+=n;else r+=n;a++}this.rawRegex=t,this.cleanedRegex=r,this.regex=new RegExp(this.cleanedRegex,"g"+e.replace("g","")),this.mapping=o}return t.prototype.regex=null,t.prototype.rawRegex=null,t.prototype.cleanedRegex=null,t.prototype.mapping=null,t.prototype.exec=function(t){var e,n,s,r;if(this.regex.lastIndex=0,n=this.regex.exec(t),null==n)return null;if(null!=this.mapping)for(s in r=this.mapping,r)e=r[s],n[s]=n[e];return n},t.prototype.test=function(t){return this.regex.lastIndex=0,this.regex.test(t)},t.prototype.replace=function(t,e){return this.regex.lastIndex=0,t.replace(this.regex,e)},t.prototype.replaceAll=function(t,e,n){var s;null==n&&(n=0),this.regex.lastIndex=0,s=0;while(this.regex.test(t)&&(0===n||s<n))this.regex.lastIndex=0,t=t.replace(this.regex,e),s++;return[t,s]},t}(),t.exports=e}}]);