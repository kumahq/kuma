(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["circuit-breakers~dataplanes~dataplanes-gateway~dataplanes-ingress~dataplanes-standard~fault-injectio~243acab4"],{"0ada":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"label-list"},[t.isReady?a("div",{staticClass:"label-list-content"},[t.isLoading||t.isEmpty?t._e():a("KCard",{attrs:{"border-variant":"noBorder"}},[a("template",{slot:"body"},[a("div",{staticClass:"label-list__col-wrapper multi-col"},[t._t("default")],2)])],2)],1):t._e(),t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")])],2):t._e(),t.isEmpty&&!t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" There is no data to display. ")])],2):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")])],2):t._e()],1)},s=[],i={name:"LabelList",props:{items:{type:Object,default:null},title:{type:String,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},computed:{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading}}},r=i,o=(a("d835"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,null,null);e["a"]=l.exports},"0aff":function(t,e,a){"use strict";var n=a("50c5"),s=a.n(n);s.a},"10d5":function(t,e,a){"use strict";var n=a("b006"),s=a.n(n);s.a},1799:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"pagination"},[t.hasPrevious?a("KButton",{ref:"paginatePrev",attrs:{appearance:"primary"},on:{click:function(e){return t.$emit("previous")}}},[t._v(" ‹ Previous ")]):t._e(),t.hasNext?a("KButton",{ref:"paginateNext",attrs:{appearance:"primary"},on:{click:function(e){return t.$emit("next")}}},[t._v(" Next › ")]):t._e()],1)},s=[],i={name:"Pagination",props:{hasPrevious:{type:Boolean,default:!1},hasNext:{type:Boolean,default:!1}}},r=i,o=(a("8b0c"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,"83b42e0c",null);e["a"]=l.exports},"1d10":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"component-frame"},[t._t("default")],2)},s=[],i={name:"FrameSkeleton"},r=i,o=(a("8463"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,"664e217a",null);e["a"]=l.exports},"23d6":function(t,e,a){"use strict";var n=a("b91b"),s=a.n(n);s.a},"251b":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"tab-container"},[t.$slots.tabHeader&&t.isReady?a("header",{staticClass:"tab__header"},[t._t("tabHeader")],2):t._e(),a("div",{staticClass:"tab__content-container",class:{"has-border":t.hasBorder}},[t.isReady?a("KTabs",{key:t.activeTab,attrs:{tabs:t.tabs},on:{changed:function(e){return t.switchTab(e)}},model:{value:t.activeTab,callback:function(e){t.activeTab=e},expression:"activeTab"}},[t._l(t.tabs,(function(e){return a("template",{slot:e.hash.replace("#","")},[t._t(e.hash.replace("#",""))],2)}))],2):t._e(),!0===t.loaders?a("div",[t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")])],2):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")])],2):t._e()],1):t._e()],1)])},s=[],i=a("ad12"),r=a.n(i),o={name:"Tabs",components:{KEmptyState:r.a},props:{loaders:{type:Boolean,default:!0},vuexState:{type:String,default:"updateSelectedTab"},isLoading:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},tabs:{type:Array,required:!0},tabGroupTitle:{type:String,default:null},hasBorder:{type:Boolean,default:!1},tabState:{type:String,default:null},initialTabOverride:{type:String,default:null}},computed:{activeTab:{get:function(){return this.tabState?"#".concat(this.$store.state[this.tabState]):this.$store.state.selectedTab},set:function(t){return t}},isReady:function(){return!1===this.loaders||!this.isEmpty&&!this.hasError&&!this.isLoading}},beforeMount:function(){this.$store.dispatch(this.vuexState,"#".concat(this.initialTabOverride)||!1)},methods:{switchTab:function(t){this.activeTab=t,this.$store.dispatch(this.vuexState,t)}}},l=o,c=(a("3be0"),a("0aff"),a("2877")),u=Object(c["a"])(l,n,s,!1,null,"ea2c4884",null);e["a"]=u.exports},2778:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"data-overview"},[n("div",{staticClass:"data-table-controls mb-2"},[t._t("additionalControls"),t.displayRefreshControl?n("KButton",{staticClass:"ml-2 refresh-button",attrs:{appearance:"primary",size:"small",disabled:t.isLoading},on:{click:function(e){return t.$emit("reloadData")}}},[n("div",{staticClass:"refresh-icon",class:{"is-spinning":t.isLoading}},[n("svg",{attrs:{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 36 36"}},[n("g",{attrs:{fill:"#fff","fill-rule":"nonzero"}},[n("path",{attrs:{d:"M18 5.5a12.465 12.465 0 00-8.118 2.995 1.5 1.5 0 001.847 2.363l.115-.095A9.437 9.437 0 0118 8.5l.272.004a9.487 9.487 0 019.07 7.75l.04.246H25a.5.5 0 00-.416.777l4 6a.5.5 0 00.832 0l4-6 .04-.072A.5.5 0 0033 16.5h-2.601l-.017-.15C29.567 10.2 24.294 5.5 18 5.5zM2.584 18.723l-.04.072A.5.5 0 003 19.5h2.6l.018.15C6.433 25.8 11.706 30.5 18 30.5c3.013 0 5.873-1.076 8.118-2.995a1.5 1.5 0 00-1.847-2.363l-.115.095A9.437 9.437 0 0118 27.5l-.272-.004a9.487 9.487 0 01-9.07-7.75l-.041-.246H11a.5.5 0 00.416-.777l-4-6a.5.5 0 00-.832 0l-4 6z"}})])])]),n("span",[t._v("Refresh")])]):t._e()],2),t.isReady?n("div",{staticClass:"data-overview-content"},[!t.isLoading&&t.displayMetrics&&t.metricsData?n("MetricGrid",{attrs:{metrics:t.metricsData}}):t._e(),t.displayDataTable&&!t.tableDataIsEmpty&&t.tableData?n("div",{staticClass:"data-overview-table"},[n("KTable",{staticClass:"micro-table",class:{"data-table-is-hidden":t.tableDataIsEmpty,"has-border":t.tableHasBorder},attrs:{options:t.tableDataFiltered,"has-side-border":!1,"has-hover":"","is-clickable":""},on:{"row:click":t.tableRowHandler},scopedSlots:t._u([t.displayTableDataStatus?{key:"status",fn:function(e){var a=e.rowValue;return[n("div",{staticClass:"entity-status",class:{"is-offline":"offline"===a.toString().toLowerCase()||!1===a}},[n("span",{staticClass:"entity-status__dot"}),n("span",{staticClass:"entity-status__label"},[t._v(t._s(a))])])]}}:null,{key:"tags",fn:function(e){var a=e.rowValue;return t._l(a,(function(e,a){return n("span",{key:a,staticClass:"entity-tags",class:"entity-tags--"+a},[n("span",{staticClass:"entity-tags__label",class:"entity-tags__label--"+t.cleanTagLabel(e.label)},[t._v(" "+t._s(e.label)+" ")]),n("span",{staticClass:"entity-tags__value",class:"entity-tags__value--"+e.value},[t._v(" "+t._s(e.value)+" ")])])}))}},{key:"totalUpdates",fn:function(e){var a=e.row;return[n("span",{staticClass:"entity-total-updates"},[n("span",[t._v(" "+t._s(a.totalUpdates)+" ")])])]}},{key:"actions",fn:function(e){var a=e.row;return[t.tableDataFunctionText?n("a",{staticClass:"data-table-action-link",class:{"is-active":t.$store.state.selectedTableRow===a.name}},[t.$store.state.selectedTableRow===a.name?n("span",{staticClass:"action-link__active-state"},[t._v(" ✓ "),n("span",{staticClass:"sr-only"},[t._v(" Selected ")])]):n("span",{staticClass:"action-link__normal-state"},[t._v(" "+t._s(t.tableDataFunctionText)+" ")])]):t._e()]}}],null,!0)}),t._t("pagination")],2):t._e(),t.displayDataTable&&t.tableDataIsEmpty&&t.tableData?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("img",{attrs:{src:a("a448")}})]),t.emptyState.title?n("span",[t._v(" "+t._s(t.emptyState.title)+" ")]):n("span",[t._v(" No Items Found ")])]),t.emptyState.message?n("template",{slot:"message"},[t._v(" "+t._s(t.emptyState.message)+" ")]):t._e()],2):t._e(),t.$slots.content?n("div",{staticClass:"data-overview-content mt-6"},[t._t("content")],2):t._e()],1):t._e(),t.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.25)",size:"42"}})],1),t._v(" Data Loading… ")])],2):t._e(),t.hasError?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")])],2):t._e()],1)},s=[],i=(a("a9e3"),a("4fad"),a("ac1f"),a("5319"),a("2909")),r=a("be10"),o={name:"DataOverview",components:{MetricGrid:r["a"]},props:{pageSize:{type:Number,default:12},displayMetrics:{type:Boolean,default:!1},metricsData:{type:Array,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},emptyState:{type:Object,default:null},ctaAction:{type:Object,default:function(){}},showCta:{type:Boolean,default:!0},displayDataTable:{type:Boolean,default:!1},tableData:{type:Object,default:null},tableHasBorder:{type:Boolean,required:!1,default:!1},tableDataIsEmpty:{type:Boolean,default:!1},tableDataActionsLink:{type:String,default:null},tableActionsRouteName:{type:String,default:null},displayTableDataStatus:{type:Boolean,default:!0},displayRefreshControl:{type:Boolean,default:!0},tableDataRow:{type:String,required:!1,default:"name"},tableDataFunctionText:{type:String,required:!1,default:null}},computed:{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},tableRowCount:function(){return Object.entries(this.tableData.data).length},pageCount:function(){var t=Object.entries(this.tableData.data).length,e=this.pageSize;return Math.ceil(t/e)},tableDataFiltered:function(){var t=this.tableData.data,e=this.tableData.headers,a={headers:e,data:Object(i["a"])(t)};return a}},methods:{tableRowHandler:function(t,e,a){this.$emit("tableAction",e)},cleanTagLabel:function(t){return t.toLowerCase().replace(".","-").replace("/","-")}}},l=o,c=(a("9947"),a("2877")),u=Object(c["a"])(l,n,s,!1,null,null,null);e["a"]=u.exports},"2ccd":function(t,e,a){"use strict";var n=a("adf5"),s=a.n(n);s.a},"3be0":function(t,e,a){"use strict";var n=a("7b7d"),s=a.n(n);s.a},"4fad":function(t,e,a){var n=a("23e7"),s=a("6f53").entries;n({target:"Object",stat:!0},{entries:function(t){return s(t)}})},"50c5":function(t,e,a){},"62e5":function(t,e){var a;a=function(){function t(t,e){var a,n,s,i,r,o,l,c,u;null==e&&(e=""),s="",r=t.length,o=null,n=0,i=0;while(i<r){if(a=t.charAt(i),"\\"===a)s+=t.slice(i,+(i+1)+1||9e9),i++;else if("("===a)if(i<r-2)if(c=t.slice(i,+(i+2)+1||9e9),"(?:"===c)i+=2,s+=c;else if("(?<"===c){n++,i+=2,l="";while(i+1<r){if(u=t.charAt(i+1),">"===u){s+="(",i++,l.length>0&&(null==o&&(o={}),o[l]=n);break}l+=u,i++}}else s+=a,n++;else s+=a;else s+=a;i++}this.rawRegex=t,this.cleanedRegex=s,this.regex=new RegExp(this.cleanedRegex,"g"+e.replace("g","")),this.mapping=o}return t.prototype.regex=null,t.prototype.rawRegex=null,t.prototype.cleanedRegex=null,t.prototype.mapping=null,t.prototype.exec=function(t){var e,a,n,s;if(this.regex.lastIndex=0,a=this.regex.exec(t),null==a)return null;if(null!=this.mapping)for(n in s=this.mapping,s)e=s[n],a[n]=a[e];return a},t.prototype.test=function(t){return this.regex.lastIndex=0,this.regex.test(t)},t.prototype.replace=function(t,e){return this.regex.lastIndex=0,t.replace(this.regex,e)},t.prototype.replaceAll=function(t,e,a){var n;null==a&&(a=0),this.regex.lastIndex=0,n=0;while(this.regex.test(t)&&(0===a||n<a))this.regex.lastIndex=0,t=t.replace(this.regex,e),n++;return[t,n]},t}(),t.exports=a},"6d8a":function(t,e,a){var n,s;s=a("62e5"),n=function(){var t;function e(){}return e.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(t=String.fromCharCode)(133),t(160),t(8232),t(8233)],e.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],e.MAPPING_ESCAPEES_TO_ESCAPED=function(){var t,a,n,s;for(n={},t=a=0,s=e.LIST_ESCAPEES.length;0<=s?a<s:a>s;t=0<=s?++a:--a)n[e.LIST_ESCAPEES[t]]=e.LIST_ESCAPED[t];return n}(),e.PATTERN_CHARACTERS_TO_ESCAPE=new s("[\\x00-\\x1f]|Â|Â |â¨|â©"),e.PATTERN_MAPPING_ESCAPEES=new s(e.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),e.PATTERN_SINGLE_QUOTING=new s("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),e.requiresDoubleQuoting=function(t){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(t)},e.escapeWithDoubleQuotes=function(t){var e;return e=this.PATTERN_MAPPING_ESCAPEES.replace(t,function(t){return function(e){return t.MAPPING_ESCAPEES_TO_ESCAPED[e]}}(this)),'"'+e+'"'},e.requiresSingleQuoting=function(t){return this.PATTERN_SINGLE_QUOTING.test(t)},e.escapeWithSingleQuotes=function(t){return"'"+t.replace(/'/g,"''")+"'"},e}(),t.exports=n},7927:function(t,e,a){},"7b7d":function(t,e,a){},8218:function(t,e,a){"use strict";a("b0c0");e["a"]={methods:{sortEntities:function(t){var e=t.sort((function(t,e){return t.name>e.name||t.name===e.name&&t.mesh>e.mesh?1:-1}));return e}}}},"82c6":function(t,e,a){},8463:function(t,e,a){"use strict";var n=a("7927"),s=a.n(n);s.a},"8b0c":function(t,e,a){"use strict";var n=a("a54e"),s=a.n(n);s.a},9947:function(t,e,a){"use strict";var n=a("82c6"),s=a.n(n);s.a},a448:function(t,e){t.exports="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0OCIgaGVpZ2h0PSI0MiI+CiAgPHBhdGggZmlsbD0iI0Q5RDlEOSIgZmlsbC1ydWxlPSJldmVub2RkIiBkPSJNNDggNHY1aC0yVjZIMnYzSDBWMkMwIC44OTU0MzA1Ljg5NTQzMSAwIDIgMGg0NGMxLjEwNDU2OSAwIDIgLjg5NTQzMDUgMiAydjJ6bS0yIDI2aC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDMwdjNIMHYtOGgydjNoNHYySDJ6bTQ0LTEyaC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDE4djNIMHYtOGgydjNoNHYySDJ6bTgtMmg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTEwIDI4aDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptMTIgMTRoLTR2LTJoNHYtM2gydjNjMCAxLjEwNDU2OTUtLjg5NTQzMSAyLTIgMnpNMiA0MGg0djJIMmMtMS4xMDQ1NjkgMC0yLS44OTU0MzA1LTItMnYtM2gydjN6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTIgMnYyaDQ0VjJIMnoiLz4KPC9zdmc+Cg=="},a54e:function(t,e,a){},a710:function(t,e,a){},a9e3:function(t,e,a){"use strict";var n=a("83ab"),s=a("da84"),i=a("94ca"),r=a("6eeb"),o=a("5135"),l=a("c6b6"),c=a("7156"),u=a("c04e"),d=a("d039"),p=a("7c73"),f=a("241c").f,b=a("06cf").f,m=a("9bf2").f,h=a("58a8").trim,v="Number",g=s[v],y=g.prototype,_=l(p(y))==v,C=function(t){var e,a,n,s,i,r,o,l,c=u(t,!1);if("string"==typeof c&&c.length>2)if(c=h(c),e=c.charCodeAt(0),43===e||45===e){if(a=c.charCodeAt(2),88===a||120===a)return NaN}else if(48===e){switch(c.charCodeAt(1)){case 66:case 98:n=2,s=49;break;case 79:case 111:n=8,s=55;break;default:return+c}for(i=c.slice(2),r=i.length,o=0;o<r;o++)if(l=i.charCodeAt(o),l<48||l>s)return NaN;return parseInt(i,n)}return+c};if(i(v,!g(" 0o1")||!g("0b1")||g("+0x1"))){for(var E,T=function(t){var e=arguments.length<1?0:t,a=this;return a instanceof T&&(_?d((function(){y.valueOf.call(a)})):l(a)!=v)?c(new g(C(e)),a,T):C(e)},S=n?f(g):"MAX_VALUE,MIN_VALUE,NaN,NEGATIVE_INFINITY,POSITIVE_INFINITY,EPSILON,isFinite,isInteger,isNaN,isSafeInteger,MAX_SAFE_INTEGER,MIN_SAFE_INTEGER,parseFloat,parseInt,isInteger".split(","),x=0;S.length>x;x++)o(g,E=S[x])&&!o(T,E)&&m(T,E,b(g,E));T.prototype=y,y.constructor=T,r(s,v,T)}},adf5:function(t,e,a){},b006:function(t,e,a){},b91b:function(t,e,a){},be10:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return t.metrics?a("KCard",{staticClass:"info-grid-wrapper mb-4"},[a("template",{slot:"body"},[a("div",{staticClass:"info-grid",class:t.metricCountClass},t._l(t.metrics,(function(e,n){return null!==e.value?a("div",{key:n,staticClass:"metric",class:e.status,attrs:{"data-testid":t._f("formatTestId")(e.metric)}},[e.url?a("router-link",{staticClass:"metric-card",attrs:{to:e.url}},[a("div",{staticClass:"metric-title color-black-85 font-semibold"},[t._v(" "+t._s(e.metric)+" ")]),a("span",{staticClass:"metric-value mt-2 type-xl",class:{"has-error":n===t.hasError[n],"has-extra-label":e.extraLabel}},[t._v(" "+t._s(t._f("formatError")(t._f("formatValue")(e.value)))+" "),e.extraLabel?a("em",{staticClass:"metric-extra-label"},[t._v(" "+t._s(e.extraLabel)+" ")]):t._e()])]):a("div",{staticClass:"metric-card"},[a("span",{staticClass:"metric-title"},[t._v(" "+t._s(e.metric)+" ")]),a("span",{staticClass:"metric-value",class:{"has-error":n===t.hasError[n]}},[t._v(" "+t._s(t._f("formatError")(t._f("formatValue")(e.value)))+" ")])])],1):t._e()})),0)])],2):t._e()},s=[],i=(a("4160"),a("b64b"),a("d3b7"),a("ac1f"),a("25f0"),a("5319"),a("159b"),{name:"MetricsGrid",filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},formatError:function(t){return"--"===t?"error calculating":t},formatTestId:function(t){return t.replace(" ","-").toLowerCase()}},props:{metrics:{type:Array,required:!0,default:function(){}}},computed:{hasError:function(){var t=this,e={};return Object.keys(this.metrics).forEach((function(a){"--"===t.metrics[a].value&&(e[a]=a)})),e},metricCountClass:function(){var t=this.metrics.length,e="metric-count--";return"".concat(e,t%3?"odd":"even")}}}),r=i,o=(a("10d5"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,"677acc48",null);e["a"]=l.exports},d835:function(t,e,a){"use strict";var n=a("a710"),s=a.n(n);s.a},e80b:function(t,e,a){var n=a("6d8a"),s="  ";function i(t){var e=typeof t;return t instanceof Array?"array":"string"==e?"string":"boolean"==e?"boolean":"number"==e?"number":"undefined"==e||null===t?"null":"hash"}function r(t,e){var a=i(t);switch(a){case"array":o(t,e);break;case"hash":l(t,e);break;case"string":u(t,e);break;case"null":e.push("null");break;case"number":e.push(t.toString());break;case"boolean":e.push(t?"true":"false");break}}function o(t,e){for(var a=0;a<t.length;a++){var n=t[a],i=[];r(n,i);for(var o=0;o<i.length;o++)e.push((0==o?"- ":s)+i[o])}}function l(t,e){for(var a in t){var n=[];if(t.hasOwnProperty(a)){var o=t[a];r(o,n);var l=i(o);if("string"==l||"null"==l||"number"==l||"boolean"==l)e.push(c(a)+": "+n[0]);else{e.push(c(a)+": ");for(var u=0;u<n.length;u++)e.push(s+n[u])}}}}function c(t){return t.match(/^[\w]+$/)?t:n.requiresDoubleQuoting(t)?n.escapeWithDoubleQuotes(t):n.requiresSingleQuoting(t)?n.escapeWithSingleQuotes(t):t}function u(t,e){e.push(c(t))}var d=function(t){"string"==typeof t&&(t=JSON.parse(t));var e=[];return r(t,e),e.join("\n")};t.exports=d},ff9d:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"yaml-view"},[t.isReady?a("div",{staticClass:"yaml-view-content"},[t.isLoading||t.isEmpty?t._e():a("KCard",{attrs:{title:t.title,"border-variant":"noBorder"}},[a("template",{slot:"body"},[a("KTabs",{key:t.environment,attrs:{tabs:t.tabs},model:{value:t.activeTab.hash,callback:function(e){t.$set(t.activeTab,"hash",e)},expression:"activeTab.hash"}},[a("template",{slot:"universal"},[a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var n=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"}},[a("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){n(t.yamlContent.universal)}}},[t._v(" Copy Universal YAML ")]),a("div",{attrs:{slot:"content"},slot:"content"},[a("p",[t._v("Entity copied to clipboard!")])])],1)]}}],null,!1,3426544560)}),a("prism",{staticClass:"code-block",attrs:{language:"yaml",code:t.yamlContent.universal}})],1),a("template",{slot:"kubernetes"},[a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var n=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"}},[a("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){n(t.yamlContent.kubernetes)}}},[t._v(" Copy Kubernetes YAML ")]),a("div",{attrs:{slot:"content"},slot:"content"},[a("p",[t._v("Entity copied to clipboard!")])])],1)]}}],null,!1,761844304)}),a("prism",{staticClass:"code-block",attrs:{language:"yaml",code:t.yamlContent.kubernetes}})],1)],2)],1)],2)],1):t._e(),!0===t.loaders?a("div",[t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")])],2):t._e(),t.isEmpty&&!t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" There is no data to display. ")])],2):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""}},[a("template",{slot:"title"},[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")])],2):t._e()],1):t._e()])},s=[],i=(a("b0c0"),a("4fad"),a("5530")),r=a("2f62"),o=a("2ccf"),l=a.n(o),c=(a("a878"),a("e80b")),u=a.n(c),d={name:"YamlView",components:{prism:l.a},props:{title:{type:String,default:null},content:{type:Object,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},data:function(){return{tabs:[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}]}},computed:Object(i["a"])(Object(i["a"])({},Object(r["b"])({environment:"getEnvironment"})),{},{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},activeTab:{get:function(){var t=this.environment;return{hash:"#".concat(t),nohash:t}},set:function(t){return{hash:"#".concat(t),nohash:t}}},yamlContent:function(){var t=this,e=this.content,a=function(){var e={},a=Object.assign({},t.content),n=a.name,s=a.type,i=(a.metadata,function(){var e=Object.assign({},t.content);return delete e.type,delete e.mesh,delete e.name,!!(e&&Object.entries(e).length>0)&&e});return delete a.type,delete a.name,e.apiVersion="kuma.io/v1alpha1",e.kind=s,e.metadata={name:n},i()&&(e.spec=i()),e},n={universal:u()(e),kubernetes:u()(a())};return n}})},p=d,f=(a("23d6"),a("2ccd"),a("2877")),b=Object(f["a"])(p,n,s,!1,null,"c3d9cbec",null);e["a"]=b.exports}}]);